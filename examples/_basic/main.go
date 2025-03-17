package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	httpserver "github.com/daniarmas/http"
)

func main() {
	// Context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create a new server with the given options and endpoints.
	server := httpserver.New(httpserver.Options{
		Addr:         net.JoinHostPort("0.0.0.0", "8080"),
		ReadTimeout:  5,
		WriteTimeout: 10,
		IdleTimeout:  15,
	}, httpserver.HandleFunc{
		Pattern: "/",
		Handler: PingHandler(),
	})

	// Start the server in a separate goroutine.
	go func() {
		log.Printf("Server is running at http://%s\n", server.HttpServer.Addr)
		if err := server.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Gracefully shutdown the server.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		log.Println("Shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.HttpServer.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Error shutting down server: %v", err)
		}
		log.Println("Server gracefully stopped")
	}()
	wg.Wait()
}

func PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}
}
