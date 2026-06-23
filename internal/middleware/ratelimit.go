package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// client struct holds the rate limiter configuration and the last time it was seen
type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients   = make(map[string]*client)
	mu        sync.Mutex // Prevents data races when multiple users visit concurrently
	cleanupOn bool
)

// initCleanup runs a background routine to remove stale IPs from memory
func initCleanup() {
	mu.Lock()
	if cleanupOn {
		mu.Unlock()
		return
	}
	cleanupOn = true
	mu.Unlock()

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			mu.Lock()
			for ip, client := range clients {
				// If an IP hasn't visited in 3 minutes, delete its bucket to save RAM
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// RateLimitAdmin restricts login attempts on sensitive routes
func RateLimitAdmin(next http.Handler) http.Handler {
	// Initialize background cleanup when middleware is first loaded
	initCleanup()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract the user's IP address
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		mu.Lock()
		// 2. If the IP is new, create a new rate limiter bucket for it
		if _, found := clients[ip]; !found {
			// rate.Every(2*time.Minute) = refilling speed (1 token every 2 mins)
			// 5 = Burst size (maximum capacity of the bucket / max attempts allowed upfront)
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Every(2*time.Minute), 5),
			}
		}

		clients[ip].lastSeen = time.Now()

		// 3. Check if the IP has an available token left
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			// Block them! Bucket is empty
			http.Error(w, "Too Many Requests. Please try again later.", http.StatusTooManyRequests)
			return
		}

		mu.Unlock()

		// 4. Token available! Pass request to the actual login handler
		next.ServeHTTP(w, r)
	})
}
