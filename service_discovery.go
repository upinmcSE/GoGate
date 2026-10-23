package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
)

func (s *ServerPool) startConsulWatcher(ctx context.Context, client *api.Client, serviceName string) {
	log.Printf("Starting Consul watcher for service: %s", serviceName)

	var lastIndex uint64 = 0

	go func() {
		for {
			// check context done -> cancel goroutine
			select {
			case <-ctx.Done():
				log.Printf("Consul watcher stopped for service: %s", serviceName)
				return
			default:
			}
			opts := &api.QueryOptions{
				WaitIndex: lastIndex, // Wait for changes *after* this index
				Near:      "_agent",
			}

			services, meta, err := client.Health().Service(serviceName, "", true, opts.WithContext(ctx))
			if err != nil {
				if ctx.Err() != nil {
					return // Context canceled when reload or shutdown
				}
				log.Printf("Error watching Consul service %s: %v. Retrying in 5s...", serviceName, err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Nếu không có thay đổi (timeout định kỳ của Consul query)
			if meta.LastIndex == lastIndex {
				continue
			}

			lastIndex = meta.LastIndex
			log.Printf("Consul update for %s. New index: %d. Found %d healthy instances.", serviceName, lastIndex, len(services))

			newBackendSet := make(map[string]bool)

			// Add new backends or set alive again for old backends
			for _, entry := range services {
				serviceID := entry.Service.ID
				newBackendSet[serviceID] = true

				s.mu.RLock()
				_, exists := s.backends[serviceID]
				s.mu.RUnlock()

				if !exists {
					addr := entry.Service.Address
					port := entry.Service.Port
					if addr == "" {
						addr = entry.Node.Address
					}
					serviceURL := fmt.Sprintf("http://%s:%d", addr, port)

					if err := s.AddBackend(serviceID, serviceURL); err != nil {
						log.Printf("Failed to add backend %s: %v", serviceID, err)
					}
				} else {
					// It exists, make sure it's marked as alive (in case our proxy marked it down)
					s.backends[serviceID].SetAlive(true)
				}
			}

			// Search and delete backends not exist on Consul
			s.mu.RLock()
			for serviceID := range s.backends {
				if !newBackendSet[serviceID] {
					go s.RemoveBackend(serviceID)
				}
			}
			s.mu.RUnlock()
		}
	}()
}
