package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Service struct {
	Name     string
	URL      string
	Prefixes []string
}

type Config struct {
	Port     string
	Services []Service
}

func loadConfig() Config {
	return Config{
		Port: "8080",
		Services: []Service{
			{
				Name:     "users-service",
				URL:      "http://localhost:8081",
				Prefixes: []string{"/users", "/auth"},
			},
			{
				Name:     "products-service",
				URL:      "http://localhost:8082",
				Prefixes: []string{"/products", "/categories"},
			},
		},
	}
}

func ProxyHandler(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a new client for each request to avoid keeping connections alive
		client := http.Client{
			Timeout: time.Second * 10,
		}

		backendURL := fmt.Sprintf("%s%s", service.URL, r.URL.Path)

		proxyReq, err := http.NewRequest(r.Method, backendURL, r.Body)
		if err != nil {
			http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
			log.Printf("Error creating proxy request: %v", err)
			return
		}

		for name, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}

		proxyReq.Header.Set("X-API-Gateway", "go-gateway")

		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, "Error forwarding request", http.StatusBadGateway)
			log.Printf("Error forwarding request: %v", err)
			return
		}
		defer resp.Body.Close()

		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		w.WriteHeader(resp.StatusCode)

		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				_, writeErr := w.Write(buf[:n])
				if writeErr != nil {
					log.Printf("Error writing response: %v", writeErr)
					return
				}
			}
			if err != nil {
				break
			}
		}
	}
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log after the request is done
		log.Printf(
			"Method: %s | Path: %s | Client: %s | Duration: %v",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			time.Since(startTime),
		)
	})
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":    "UP",
		"timestamp": time.Now().String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	config := loadConfig()

	router := mux.NewRouter()

	router.Use(LoggingMiddleware)

	router.HandleFunc("/health", HealthCheckHandler).Methods("GET")

	// Register routes for each service
	for _, service := range config.Services {
		for _, prefix := range service.Prefixes {
			log.Printf("Registering routes: %s/* -> %s", prefix, service.URL)
			router.PathPrefix(prefix).HandlerFunc(ProxyHandler(service))
		}
	}

	// Catchall route for unmatched paths
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("No route found for: %s %s", r.Method, r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	serverAddr := fmt.Sprintf(":%s", config.Port)
	log.Printf("API Gateway starting on %s", serverAddr)

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
