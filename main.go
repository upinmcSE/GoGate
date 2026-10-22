package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	fmt.Println("Go Gate !")

	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatalf("Can't read yaml file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Wrong YAML syntax: %v", err)
	}

	serverMux := http.NewServeMux()

	for _, svc := range config.Services {
		pool := NewServerPool()

		for i, backendURL := range svc.Backends {
			serviceID := fmt.Sprintf("%s-%d", svc.Name, i+1)
			if err := pool.AddBackend(serviceID, backendURL); err != nil {
				log.Fatalf("Invalid backend URL %s: %v", backendURL, err)
			}
		}

		serverMux.Handle(svc.Path, newServiceHandler(pool))
		log.Printf("Register route successful %s (%d backends)", svc.Path, len(svc.Backends))
	}

	log.Printf("GoGate listening on :%s", config.GatewayPort)
	log.Fatal(http.ListenAndServe(":"+config.GatewayPort, serverMux))
}
