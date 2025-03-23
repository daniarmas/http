package http

import (
	"context"
	"net/http"
	"sync"
	"time"

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

// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing all open
// listeners, then closing all idle connections, and then waiting
// indefinitely for connections to return to idle and then shut down.
// If the provided context expires before the shutdown is complete,
// Shutdown returns the context's error, otherwise it returns any
// error returned from closing the [Server]'s underlying Listener(s).
//
// When Shutdown is called, [Serve], [ListenAndServe], and
// [ListenAndServeTLS] immediately return [ErrServerClosed]. Make sure the
// program doesn't exit and waits instead for Shutdown to return.
//
// Shutdown does not attempt to close nor wait for hijacked
// connections such as WebSockets. The caller of Shutdown should
// separately notify such long-lived connections of shutdown and wait
// for them to close, if desired. See [Server.RegisterOnShutdown] for a way to
// register shutdown notification functions.
//
// Once Shutdown has been called on a server, it may not be reused;
// future calls to methods such as Serve will return ErrServerClosed.
func (s *Server) Shutdown(ctx context.Context) error {
	// Gracefully shutdown the server.
	var shutdownErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdownErr = s.HttpServer.Shutdown(shutdownCtx)
	}()
	wg.Wait()
	return shutdownErr
}
