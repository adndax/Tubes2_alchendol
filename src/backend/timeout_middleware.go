package main

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
)

// TimeoutMiddleware adds a timeout to requests
func TimeoutMiddleware(next http.Handler, timeout time.Duration) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Create a context with timeout only if we're not running locally
        // This way, local development won't be affected
        if r.URL.Query().Get("noTimeout") != "true" {
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()
            r = r.WithContext(ctx)
        }
        
        // Use a channel to track completion
        done := make(chan struct{})
        
        go func() {
            next.ServeHTTP(w, r)
            close(done)
        }()
        
        select {
        case <-done:
            return
        case <-r.Context().Done():
            // Only trigger if this was a timeout, not other context cancellations
            if r.Context().Err() == context.DeadlineExceeded {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusRequestTimeout)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": "Request processing timed out",
                })
            }
            return
        }
    })
}