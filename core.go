package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type Config struct {
	GatewayPort string    `yaml:"gatewayPort"`
	Services    []Service `yaml:"services"`
}

type Service struct {
	Name     string   `yaml:"name"`
	Path     string   `yaml:"path"`
	Backends []string `yaml:"backends"`
}

type Backend struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
	isAlive      atomic.Bool
}

func (b *Backend) SetAlive(alive bool) {
	b.isAlive.Store(alive)
}

type ServerPool struct {
	backends map[string]*Backend // key: serviceId -> value: *Backend
	ids      []string            //
	current  atomic.Uint64       // atomic count requests
	mu       sync.RWMutex        // lock
}

func NewServerPool() *ServerPool {
	return &ServerPool{
		backends: make(map[string]*Backend),
		ids:      make([]string, 0),
	}
}

func (s *ServerPool) AddBackend(serviceID string, backendURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.backends[serviceID]; exists {
		return fmt.Errorf("backend with ID %s already exist", serviceID)
	}

	parsedURL, err := url.Parse(backendURL)
	if err != nil {
		return err
	}

	backend := &Backend{
		URL: parsedURL,
	}
	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	// Passive Health Check
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("Backend error: %v", e)
		backend.SetAlive(false)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	}

	backend.SetAlive(true)
	backend.ReverseProxy = proxy
	s.backends[serviceID] = backend
	s.ids = append(s.ids, serviceID)
	log.Printf("Added backend: %s, id: %s", backendURL, serviceID)
	return nil
}

func (s *ServerPool) RemoveBackend(serviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if backend, ok := s.backends[serviceID]; ok {
		delete(s.backends, serviceID)
		for i, id := range s.ids {
			if id == serviceID {
				s.ids = append(s.ids[:i], s.ids[i+1:]...)
				break
			}
		}
		log.Printf("Removed backend: %s (ID: %s)", backend.URL, serviceID)
	}
}

func (s *ServerPool) GetNextBackend() *Backend {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.ids)
	if total == 0 {
		return nil
	}

	next := s.current.Add(1)

	for i := range total {
		idx := (next + uint64(i)) % uint64(total) // modulo operator
		serviceID := s.ids[idx]
		backend := s.backends[serviceID]
		if backend.isAlive.Load() {
			return backend
		}
	}
	return nil
}

func newServiceHandler(pool *ServerPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backend := pool.GetNextBackend()
		if backend == nil {
			log.Println("No healthy backends for this service!")
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}
		log.Printf("Forwarding request to: %s", backend.URL)
		backend.ReverseProxy.ServeHTTP(w, r)
	}
}
