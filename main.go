package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v3"
)

func main() {
	fmt.Println("Go Gate !")

	consulClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}

	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatalf("Can't read yaml file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Wrong YAML syntax: %v", err)
	}

	serverMux := http.NewServeMux()
	ctx := context.Background()

	for _, svc := range config.Services {
		if svc.ConsulServiceName == "" {
			log.Printf("Skipping service %s: missing consulServiceName", svc.Name)
			continue
		}

		pool := NewServerPool()
		pool.startConsulWatcher(ctx, consulClient, svc.ConsulServiceName)

		serverMux.Handle(svc.Path, newServiceHandler(pool))
		log.Printf("Registered route: %s -> Consul Service: '%s'", svc.Path, svc.ConsulServiceName)
	}

	log.Printf("GoGate listening on :%s", config.GatewayPort)
	log.Fatal(http.ListenAndServe(":"+config.GatewayPort, serverMux))
}
