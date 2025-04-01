package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/daniarmas/clogg"
	"github.com/daniarmas/http/middleware"
)

// Server provides a convenient wrapper around the standard library's http.Server.
type Server struct {
	HttpServer *http.Server
}

// Options contains arguments to configure a Server instance.
type Options struct {
	// Addr optionally specifies the TCP address for the server to listen on,
	// in the form "host:port". If empty, ":http" (port 80) is used.
	// The service names are defined in RFC 6335 and assigned by IANA.
	// See net.Dial for details of the address format.
	Addr string

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body. A zero or negative value means
	// there will be no timeout.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	// A zero or negative value means there will be no timeout.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If zero, the value
	// of ReadTimeout is used. If negative, or if zero and ReadTimeout
	// is zero or negative, there is no timeout.
	IdleTimeout time.Duration

	// Middlewares is a slice of http.Handler that can be used to apply
	// middleware to the HTTP server. These middlewares will be applied
	// in the order they are provided.
	Middlewares []middleware.Middleware
}

// HandleFunc is a struct that contains the pattern and the handler function.
type HandleFunc struct {
	Pattern string
	Handler http.HandlerFunc
}

// NewServer creates a new Server instance with the given endpoints and options.
func New(opts Options, endpoints ...HandleFunc) *Server {
	// Create a new ServeMux and register the endpoints.
	mux := http.NewServeMux()
	// Health check endpoint
	mux.HandleFunc("GET /health", HealthCheckHandler)
	// Not found endpoint
	mux.HandleFunc("/", NotFoundHandler)
	// Register the endpoints
	for _, endpoint := range endpoints {
		mux.HandleFunc(endpoint.Pattern, endpoint.Handler)
	}

	var handler http.Handler = mux
	// Add the middlewares in reverse order so they are executed in the order they are provided.
	for i := len(opts.Middlewares) - 1; i >= 0; i-- {
		handler = opts.Middlewares[i](handler)
	}

	return &Server{
		HttpServer: &http.Server{
			Addr:         opts.Addr,
			Handler:      handler,
			ReadTimeout:  opts.ReadTimeout,
			WriteTimeout: opts.WriteTimeout,
			IdleTimeout:  opts.IdleTimeout,
		},
	}
}

// Run starts the server and handles graceful shutdown.
func (s *Server) Run(ctx context.Context) error {
	var err error

	// Start the server in a separate goroutine.
	go func() {
		clogg.Info(ctx, fmt.Sprintf("server is running at http://%s", s.HttpServer.Addr))
		if err = s.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			err = fmt.Errorf("error starting server: %w", err)
		}
	}()

	// Wait for the context to be canceled.
	<-ctx.Done()

	// Gracefully shutdown the server.
	clogg.Info(ctx, "shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = s.HttpServer.Shutdown(shutdownCtx); err != nil {
		err = fmt.Errorf("error shutting down server: %w", err)
		return err
	} else {
		clogg.Info(ctx, "server gracefully stopped")
	}

	return err
}
