package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rendyspratama/digital-discovery/api/config"
	"github.com/rendyspratama/digital-discovery/api/routes"
)

func main() {
	// ANSI color codes
	green := "\033[32m"
	blue := "\033[34m"
	reset := "\033[0m"
	bold := "\033[1m"

	// Load configuration
	cfg := config.LoadConfig()

	// Setup router
	router := routes.SetupRouter()

	// Create server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("\n%s%s=== Digital Discovery API ===%s\n", bold, blue, reset)
		fmt.Printf("\n%s▶ Server starting on port %s%s\n", green, cfg.Port, reset)
		fmt.Printf("%s▶ Time: %s%s\n", green, time.Now().Format("2006-01-02 15:04:05"), reset)
		fmt.Printf("%s▶ Environment: %s%s\n\n", green, os.Getenv("GO_ENV"), reset)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("%sServer failed to start: %v%s", bold, err, reset)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	fmt.Printf("\n%s⏹ Shutting down server...%s\n", blue, reset)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("%sServer forced to shutdown: %v%s", bold, err, reset)
	}

	fmt.Printf("%s✓ Server exited properly%s\n\n", green, reset)
}
