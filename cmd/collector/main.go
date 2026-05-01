package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/codriverlabs/KubeCoDriver/pkg/collector/server"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	dateFormat := os.Getenv("DATE_FORMAT")
	if dateFormat == "" {
		log.Fatalf("DATE_FORMAT environment variable is required")
	}

	cfg := &server.Config{
		Port:        8443,
		StoragePath: "/data",
		DateFormat:  dateFormat,
		TLSCert:     os.Getenv("TLS_CERT_PATH"),
		TLSKey:      os.Getenv("TLS_KEY_PATH"),
	}

	// Create Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to create kubernetes config: %v", err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create kubernetes client: %v", err)
	}

	srv, err := server.NewServer(cfg, k8sClient)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	<-sigChan
	if err := srv.Shutdown(); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}
}
