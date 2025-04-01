package middleware

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/daniarmas/clogg"
	"github.com/daniarmas/http/response"
)

type Middleware func(http.Handler) http.Handler

type CorsOptions struct {
	AllowedOrigin  string
	AllowedMethods []string
	AllowedHeaders []string
}

func (o CorsOptions) Validate() bool {
	if len(o.AllowedOrigin) != 0 && !validateCorsOrigin(o.AllowedOrigin) {
		return false
	}

	return true
}

func validateCorsOrigin(origin string) bool {
	// Define the regex pattern for a valid URL with optional port
	pattern := `^(\*|https?:\/\/(?:\d{1,3}\.){3}\d{1,3}(?::\d+)?|https?:\/\/[a-zA-Z0-9.-]+(?::\d+)?)$`

	// Compile the regex pattern
	re := regexp.MustCompile(pattern)

	// Validate the origin using the regex pattern
	return re.MatchString(origin)
}

// RecoverMiddleware is an HTTP middleware that recovers from panics and returns 500 Internal Server Error.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				clogg.Error(r.Context(), "recovered from panic", clogg.String("error", err.(string)), clogg.String("stack", string(debug.Stack())))
				response.InternalServerError(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// AllowCORS returns a middleware that sets the CORS headers.
func AllowCors(options CorsOptions) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if options.Validate() && origin == options.AllowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", options.AllowedOrigin)
				// Handle preflight requests.
				if r.Method == "OPTIONS" {
					response.NoContent(w, r)
					return
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Call the next handler.
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs the details of each API request and response.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer that captures the status code and response size.
		lrw := &loggingResponseWriter{ResponseWriter: w}

		// Call the next handler.
		next.ServeHTTP(lrw, r)

		// Extract the client IP address without the port.
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			clientIP = r.RemoteAddr
		} else {
			// Convert to IPv4 if possible
			ip := net.ParseIP(clientIP)
			if ip != nil {
				if ipv4 := ip.To4(); ipv4 != nil {
					clientIP = ipv4.String()
				}
			}
		}

		// Calculate the response time in microseconds.
		durationMicro := time.Since(start).Microseconds()
		durationMilli := float64(durationMicro) / 1000.0

		// Log the request and response details.
		clogg.Info(r.Context(), "request",
			clogg.String("endpoint", r.RequestURI),
			clogg.String("method", r.Method),
			clogg.Int("status", lrw.statusCode),
			clogg.String("duration", fmt.Sprintf("%.3f ms", durationMilli)),
			clogg.String("response_size", fmt.Sprintf("%d bytes", lrw.responseSize)),
			clogg.String("client_ip", clientIP),
			clogg.String("user_agent", r.UserAgent()),
		)
	})
}

// loggingResponseWriter is a custom response writer that captures the status code and response size.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int
}

// WriteHeader captures the status code.
func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response size.
func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.responseSize += size
	return size, err
}
