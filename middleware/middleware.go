package middleware

import (
	"net/http"
	"regexp"

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
