package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-username/virtumancer/internal/api"
	"github.com/your-username/virtumancer/internal/libvirt"
	"github.com/your-username/virtumancer/internal/services"
	"github.com/your-username/virtumancer/internal/storage"
)

// spaFileSystem is a custom file system that serves the index.html
// for any path that is not found. This is necessary for single-page applications.
type spaFileSystem struct {
	root http.FileSystem
}

func (fs spaFileSystem) Open(name string) (http.File, error) {
	f, err := fs.root.Open(name)
	if os.IsNotExist(err) {
		return fs.root.Open("index.html")
	}
	return f, err
}

func main() {
	log.Println("Starting VirtuMancer - The Ultimate Libvirt Web UI")

	// --- Core Application Services ---

	// 1. Initialize Database
	db, err := storage.NewDB("virtumancer.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection established.")

	// 2. Initialize the Libvirt Connection Manager
	connector := libvirt.NewConnector()

	// 3. Initialize Host Service (bridges API, DB, and Libvirt)
	hostService := services.NewHostService(db, connector)

	// 4. Load and connect to all saved hosts on startup
	if err := hostService.ConnectAllHosts(); err != nil {
		log.Printf("WARNING: Could not connect to one or more saved hosts: %v", err)
	}

	// 5. Initialize the API Handler with dependencies
	apiHandler := &api.APIHandler{
		LibvirtConnector: connector,
		HostService:      hostService,
	}

	// --- Router and Middleware Setup ---

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// --- API Routes ---

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", api.HealthCheck)

		// Host management routes
		r.Get("/hosts", apiHandler.ListHosts)
		r.Post("/hosts", apiHandler.AddHost)
		r.Delete("/hosts/{hostID}", apiHandler.DeleteHost)

		// VM routes (scoped by host)
		r.Get("/hosts/{hostID}/vms", apiHandler.ListVMs)
	})

	// --- Frontend File Server ---

	webDir := "./web"
	fs := http.FileServer(spaFileSystem{root: http.Dir(webDir)})
	r.Handle("/*", fs)

	// --- Server Start ---

	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("Server listening on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}


