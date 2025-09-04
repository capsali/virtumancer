package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/capsali/virtumancer/internal/api"
	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/services"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// --- Initialization ---
	// 1. Initialize the SQLite database connection.
	db, err := storage.InitDB("virtumancer.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 2. Create the libvirt connector which manages all host connections.
	connector := libvirt.NewConnector()

	// 3. Create the host service, injecting the database and connector.
	// This service contains the core business logic.
	hostService := services.NewHostService(db, connector)

	// 4. On startup, load all hosts from the DB and attempt to connect to them.
	// This ensures that previously added hosts are available after a restart.
	hostService.ConnectToAllHosts()

	// 5. Create the API handler, injecting the host service.
	// This handler exposes the service's functionality via HTTP endpoints.
	apiHandler := api.NewAPIHandler(hostService)

	// --- Router Setup ---
	// Create a new Chi router and add essential middleware.
	r := chi.NewRouter()
	r.Use(middleware.Logger)   // Log requests to the console.
	r.Use(middleware.Recoverer) // Recover from panics without crashing the server.

	// --- API Routes ---
	// Group all API endpoints under the /api/v1 prefix.
	r.Route("/api/v1", func(r chi.Router) {
		// Host management endpoints
		r.Get("/health", apiHandler.HealthCheck)
		r.Get("/hosts", apiHandler.GetHosts)
		r.Post("/hosts", apiHandler.CreateHost)
		r.Delete("/hosts/{hostID}", apiHandler.DeleteHost)

		// VM management endpoints
		r.Get("/hosts/{hostID}/vms", apiHandler.ListVMs)
		r.Post("/hosts/{hostID}/vms/{vmName}/start", apiHandler.StartVM)
		r.Post("/hosts/{hostID}/vms/{vmName}/graceful-shutdown", apiHandler.GracefulShutdownVM)
		r.Post("/hosts/{hostID}/vms/{vmName}/graceful-reboot", apiHandler.GracefulRebootVM)
		r.Post("/hosts/{hostID}/vms/{vmName}/force-off", apiHandler.ForceOffVM)
		r.Post("/hosts/{hostID}/vms/{vmName}/force-reset", apiHandler.ForceResetVM)
	})

	// --- Static File Server ---
	// This section serves the compiled Vue.js frontend.
	workDir, _ := os.Getwd()
	staticFilesPath := workDir + "/web/dist"
	fileServer := http.FileServer(http.Dir(staticFilesPath))

	// Handle all other requests by serving the frontend.
	// This includes a fallback to index.html for Single-Page Application (SPA) routing.
	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		filePath := staticFilesPath + r.URL.Path
		_, err := os.Stat(filePath)
		// If the file doesn't exist, it's a client-side route; serve index.html.
		if os.IsNotExist(err) || strings.HasSuffix(r.URL.Path, "/") {
			http.ServeFile(w, r, staticFilesPath+"/index.html")
		} else {
			// Otherwise, serve the static file (e.g., CSS, JS).
			fileServer.ServeHTTP(w, r)
		}
	})

	// --- Start Server ---
	log.Println("Starting VirtuMancer server on http://localhost:8888")
	err = http.ListenAndServe(":8888", r)
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}


