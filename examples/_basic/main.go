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

	// Routes
	routes := []httpserver.HandleFunc{
		{
			Pattern: "/panic",
			Handler: PanicHandler,
		},
		{
			Pattern: "/ping",
			Handler: PingHandler,
		},
	}

	// Create a new server with the given options and endpoints.
	server := httpserver.New(httpserver.Options{
		Addr:         net.JoinHostPort("0.0.0.0", "8080"),
		ReadTimeout:  1000 * time.Second, // 1000 seconds to allow for debugging
		WriteTimeout: 1000 * time.Second, // 1000 seconds to allow for debugging
		IdleTimeout:  1000 * time.Second, // 1000 seconds to allow for debugging
		Middlewares: []middleware.Middleware{
			middleware.LoggingMiddleware,
			middleware.AllowCors(middleware.CorsOptions{
				AllowedOrigin:  "http://localhost:3000",
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{"Content-Type", "Authorization"},
			}),
			middleware.RecoverMiddleware,
		},
	}, routes...)

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

// PanicHandler is an HTTP handler that panics.
func PanicHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate a panic to test the RecoverMiddleware.
	panic("Oops!")
}

// PingHandler is an HTTP handler that returns a pong response.
func PingHandler(w http.ResponseWriter, r *http.Request) {
	response.OK(w, r, "pong")
}
