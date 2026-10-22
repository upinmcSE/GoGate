package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GatewayPort string    `yaml:"gatewayPort"`
	Services    []Service `yaml:"services"`
}
type Service struct {
	Name       string `yaml:"name"`
	Path       string `yaml:"path"`
	BackendURL string `yaml:"backendURL"`
}

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
		backendUrl, err := url.Parse(svc.BackendURL)
		if err != nil {
			log.Fatalf("Invalid URL %s: %v", svc.Name, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(backendUrl)

		serverMux.Handle(svc.Path, proxy)

		log.Printf("Register route successful %s -> %s", svc.Path, svc.BackendURL)
	}

	log.Printf("GoGate listening on :%s", config.GatewayPort)
	log.Fatal(http.ListenAndServe(":"+config.GatewayPort, serverMux))
}
