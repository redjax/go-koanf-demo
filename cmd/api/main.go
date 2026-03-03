package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourname/go-conf-demo/internal/config"
)

var appConfig *config.Config

func main() {
	// Load configuration
	configFile := "config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	var err error
	appConfig, err = config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup HTTP server using config values
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/config", configHandler)
	mux.HandleFunc("/", rootHandler)

	addr := fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  appConfig.Server.ReadTimeout,
		WriteTimeout: appConfig.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting API server on %s", addr)
		log.Printf("Environment: %s", appConfig.App.Environment)
		log.Printf("Debug mode: %t", appConfig.App.Debug)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown using config timeout
	ctx, cancel := context.WithTimeout(context.Background(), appConfig.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":      "healthy",
		"timestamp":   time.Now().Format(time.RFC3339),
		"environment": appConfig.App.Environment,
		"version":     appConfig.App.Version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	// Return sanitized config (no passwords)
	response := map[string]interface{}{
		"app": map[string]interface{}{
			"name":        appConfig.App.Name,
			"version":     appConfig.App.Version,
			"environment": appConfig.App.Environment,
			"debug":       appConfig.App.Debug,
		},
		"server": map[string]interface{}{
			"host": appConfig.Server.Host,
			"port": appConfig.Server.Port,
		},
		"database": map[string]interface{}{
			"host": appConfig.Database.Host,
			"port": appConfig.Database.Port,
			"name": appConfig.Database.Name,
			// Intentionally omit password
		},
		"features": map[string]interface{}{
			"metrics":     appConfig.Features.EnableMetrics,
			"tracing":     appConfig.Features.EnableTracing,
			"rate_limits": appConfig.Features.EnableRateLimits,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	response := map[string]interface{}{
		"message": "Go Config Demo API",
		"version": appConfig.App.Version,
		"endpoints": []string{
			"/health - Health check",
			"/config - View configuration",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
