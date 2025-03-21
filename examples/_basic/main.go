package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpserver "github.com/daniarmas/http"

	"github.com/daniarmas/http/middleware"
	"github.com/daniarmas/http/response"
)

func main() {
	// Context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create a new server with the given options and endpoints.
	server := httpserver.New(httpserver.Options{
		Addr:         net.JoinHostPort("0.0.0.0", "8080"),
		ReadTimeout:  1000 * time.Second,
		WriteTimeout: 1000 * time.Second,
		IdleTimeout:  1500 * time.Second,
		Middlewares: []middleware.Middleware{
			middleware.AllowCors(middleware.CorsOptions{
				AllowedOrigin:  "http://localhost:3000",
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{"Content-Type", "Authorization"},
			}),
		},
	}, httpserver.HandleFunc{
		Pattern: "/ping",
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
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("error shutting down server: %v", err)
	}
}

func PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, r, "pong")
	}
}
