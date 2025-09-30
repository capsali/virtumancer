package main

import (
	"net/http"
	"os"

	log "github.com/capsali/virtumancer/internal/logging"

	"github.com/capsali/virtumancer/internal/api"
	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/services"
	"github.com/capsali/virtumancer/internal/storage"
	"github.com/capsali/virtumancer/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Parse logging flags early
	var verboseFlag bool
	var debugFlag bool
	for _, a := range os.Args[1:] {
		switch a {
		case "--verbose":
			verboseFlag = true
		case "--debug":
			debugFlag = true
		}
	}
	// Default to info if none specified
	if debugFlag {
		log.SetLevel("debug")
	} else if verboseFlag {
		log.SetLevel("verbose")
	} else {
		log.SetLevel("info")
	}
	// Initialize Database
	db, err := storage.InitDB("virtumancer.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// Initialize Libvirt Connector
	connector := libvirt.NewConnector()

	// Initialize Host Service
	hostService := services.NewHostService(db, connector, hub)

	// Auto-connect to hosts that were previously connected
	if err := hostService.AutoConnectHosts(); err != nil {
		log.Errorf("Failed to auto-connect to some hosts: %v", err)
	}

	// Host connections are established lazily when needed (e.g., on the
	// first websocket subscription) to avoid delaying server startup.

	// Initialize API Handler
	apiHandler := api.NewAPIHandler(hostService, hub, db, connector)

	// Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", apiHandler.HealthCheck)

		// Host routes
		r.Get("/hosts", apiHandler.GetHosts)
		r.Post("/hosts", apiHandler.CreateHost)

		// Global discovered VMs routes
		r.Get("/discovered-vms", apiHandler.ListAllDiscoveredVMs)
		r.Post("/discovered-vms/refresh", apiHandler.RefreshAllDiscoveredVMs)

		r.Post("/hosts/{hostID}/connect", apiHandler.ConnectHost)
		r.Post("/hosts/{hostID}/disconnect", apiHandler.DisconnectHost)
		r.Get("/hosts/{hostID}/info", apiHandler.GetHostInfo)
		r.Get("/hosts/{hostID}/stats", apiHandler.GetHostStats)
		r.Patch("/hosts/{hostID}", apiHandler.UpdateHost)
		r.Delete("/hosts/{hostID}", apiHandler.DeleteHost)

		// VM routes
		r.Get("/hosts/{hostID}/vms", apiHandler.ListVMsFromLibvirt)
		// Discovered/Import routes
		r.Get("/hosts/{hostID}/discovered-vms", apiHandler.ListDiscoveredVMs)
		r.Post("/hosts/{hostID}/vms/{vmName}/import", apiHandler.ImportVM)
		r.Post("/hosts/{hostID}/vms/import-all", apiHandler.ImportAllVMs)
		r.Post("/hosts/{hostID}/vms/import-selected", apiHandler.ImportSelectedVMs)
		r.Delete("/hosts/{hostID}/discovered-vms", apiHandler.DeleteSelectedDiscoveredVMs)
		r.Post("/hosts/{hostID}/vms/{vmName}/start", apiHandler.StartVM)
		r.Post("/hosts/{hostID}/vms/{vmName}/shutdown", apiHandler.ShutdownVM)
		r.Post("/hosts/{hostID}/vms/{vmName}/reboot", apiHandler.RebootVM)
		r.Post("/hosts/{hostID}/vms/{vmName}/forceoff", apiHandler.ForceOffVM)
		r.Post("/hosts/{hostID}/vms/{vmName}/forcereset", apiHandler.ForceResetVM)
		r.Post("/hosts/{hostID}/vms/{vmName}/sync-from-libvirt", apiHandler.SyncVMLive)
		r.Post("/hosts/{hostID}/vms/{vmName}/rebuild-from-db", apiHandler.RebuildVM)
		r.Put("/hosts/{hostID}/vms/{vmName}/state", apiHandler.UpdateVMState)
		r.Get("/hosts/{hostID}/vms/{vmName}/stats", apiHandler.GetVMStats)
		r.Get("/hosts/{hostID}/vms/{vmName}/hardware", apiHandler.GetVMHardware)
		r.Get("/hosts/{hostID}/vms/{vmName}/hardware/extended", apiHandler.GetVMExtendedHardware)

		// Port routes
		r.Get("/hosts/{hostID}/ports", apiHandler.ListHostPorts)
		r.Get("/hosts/{hostID}/vms/{vmName}/port-attachments", apiHandler.ListVMPortAttachments)

		// Storage routes
		r.Get("/storage/pools", apiHandler.ListStoragePools)
		r.Get("/storage/volumes", apiHandler.ListStorageVolumes)
		r.Get("/storage/disk-attachments", apiHandler.ListDiskAttachments)
		r.Get("/hosts/{hostID}/storage/pools", apiHandler.ListHostStoragePools)
		r.Get("/hosts/{hostID}/storage/volumes", apiHandler.ListHostStorageVolumes)

		// Network routes
		r.Get("/networks", apiHandler.ListNetworks)
		r.Get("/ports", apiHandler.ListPorts)
		r.Get("/port-attachments", apiHandler.ListPortAttachments)
		r.Get("/hosts/{hostID}/networks", apiHandler.ListHostNetworks)

		// Video / GPU routes
		r.Get("/video/models", apiHandler.ListVideoModels)
		r.Get("/hosts/{hostID}/video/devices", apiHandler.ListHostVideoDevices)
		r.Get("/hosts/{hostID}/vms/{vmName}/video-attachments", apiHandler.ListVMVideoAttachments)

		// Console routes
		r.Get("/hosts/{hostID}/vms/{vmName}/console", apiHandler.HandleVMConsole)
		r.Get("/hosts/{hostID}/vms/{vmName}/spice", apiHandler.HandleSpiceConsole)

		// Dashboard routes
		r.Get("/dashboard/stats", apiHandler.GetDashboardStats)
		r.Get("/dashboard/activity", apiHandler.GetDashboardActivity)
		r.Get("/dashboard/overview", apiHandler.GetDashboardOverview)

		// Settings routes
		r.Get("/settings/metrics", apiHandler.GetMetricsSettings)
		r.Put("/settings/metrics", apiHandler.UpdateMetricsSettings)
		r.Get("/settings/metrics/runtime", apiHandler.GetRuntimeMetricsSettings)
	})

	// WebSocket route for UI updates
	r.HandleFunc("/ws", apiHandler.HandleWebSocket)

	// Static File Server for the Vue App
	workDir, _ := os.Getwd()

	spiceDir := http.Dir(workDir + "/web/public/spice")
	r.Handle("/spice/*", http.StripPrefix("/spice/", http.FileServer(spiceDir)))

	fileServer := http.FileServer(http.Dir(workDir + "/web/dist"))
	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		_, err := os.Stat(workDir + "/web/dist" + r.URL.Path)
		if os.IsNotExist(err) {
			http.ServeFile(w, r, workDir+"/web/dist/index.html")
		} else {
			fileServer.ServeHTTP(w, r)
		}
	})

	certFile := "localhost.crt"
	keyFile := "localhost.key"

	log.Infof("Starting HTTPS server on :8890")
	err = http.ListenAndServeTLS(":8890", certFile, keyFile, r)
	if err != nil {
		log.Debugf("Could not start HTTPS server: %v", err)
		log.Infof("Please ensure 'localhost.crt' and 'localhost.key' are present in the root directory.")
		log.Infof("You can generate them by running the 'generate-certs.sh' script.")
	}
}
