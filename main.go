package main

import (
	"log"
	"net/http"
	"os"

	"github.com/capsali/virtumancer/internal/api"
	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/services"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Initialize Database
	db, err := storage.InitDB("virtumancer.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Libvirt Connector
	connector := libvirt.NewConnector()

	// Initialize Host Service
	hostService := services.NewHostService(db, connector)

	// On startup, load all hosts from DB and try to connect
	hostService.ConnectToAllHosts() // This function logs its own errors

	// Initialize API Handler
	apiHandler := api.NewAPIHandler(hostService)

	// Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", apiHandler.HealthCheck)
		r.Get("/hosts", apiHandler.GetHosts)
		r.Post("/hosts", apiHandler.CreateHost)
		r.Delete("/hosts/{hostID}", apiHandler.DeleteHost)
		r.Get("/hosts/{hostID}/vms", apiHandler.ListVMs)
	})

	// --- Static File Server ---
	// Get the current working directory
	workDir, _ := os.Getwd()
	// Create a file system from the 'web/dist' directory
	fileServer := http.FileServer(http.Dir(workDir + "/web/dist"))

	// Serve static files, but handle SPA routing
	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		// Check if the requested file exists
		_, err := os.Stat(workDir + "/web/dist" + r.URL.Path)
		// If it doesn't exist, serve index.html for SPA routing
		if os.IsNotExist(err) {
			http.ServeFile(w, r, workDir+"/web/dist/index.html")
		} else {
			// Otherwise, serve the static file
			fileServer.ServeHTTP(w, r)
		}
	})

	log.Println("Starting server on :8888")
	err = http.ListenAndServe(":8888", r)
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}


