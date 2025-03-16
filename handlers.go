package http

import "net/http"

// HealthCheckResponse represents the structure of the health check response
type HealthCheckResponse struct {
	Status string `json:"status"`
}

// HealthCheckHandler handles HTTP requests for checking the health status of the server.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	res := HealthCheckResponse{Status: "healthy"}
	OK(w, r, res)
}
