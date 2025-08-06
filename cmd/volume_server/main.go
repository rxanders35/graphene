package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/rxanders35/sss/volume_server"
)

func main() {
	mastergRPCAddr := flag.String("master-addr", "localhost:9090", "master's grpc address")
	volumeHTTPAddr := flag.String("volume-addr", "localhost:8080", "volume's http address")
	dataDir := flag.String("data-dir", "./data", "volume's data namespace")

	flag.Parse()

	serverId, err := getOrCreateServerID(*dataDir)
	if err != nil {
		log.Fatal(err)
	}

	v, err := volume_server.NewVolume(*dataDir, serverId)
	if err != nil {
		log.Fatalf("Couldn't init volume backend. Why: %v", err)
	}

	m := volume_server.NewMasterClient(*mastergRPCAddr)

	s, err := volume_server.NewHTTPServer(*volumeHTTPAddr, v, m, serverId)
	if err != nil {
		log.Fatalf("Couldn't init volume server. Why: %v", err)
	}

	go func() {
		if err := s.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server run error. Why: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shut down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed. Why: %v", err)
	}
}

func getOrCreateServerID(dataDir string) (uuid.UUID, error) {
	volumeServerIdPath := dataDir + "/volume.id"

	// Check if the path exists and is a directory
	if info, err := os.Stat(volumeServerIdPath); err == nil && info.IsDir() {
		return uuid.Nil, fmt.Errorf("volume.id path %s is a directory, expected a file", volumeServerIdPath)
	}

	// Try reading the file
	idData, err := os.ReadFile(volumeServerIdPath)
	if err == nil {
		volumeServerId, err := uuid.FromBytes(idData)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed parsing volume server id from %s: %v", volumeServerIdPath, err)
		}
		return volumeServerId, nil
	}

	// If the file doesn't exist, create a new UUID and write it
	if os.IsNotExist(err) {
		volumeServerId := uuid.New()
		// Ensure the directory exists
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return uuid.Nil, fmt.Errorf("failed to create directory %s: %v", dataDir, err)
		}
		// Write the UUID to the file
		if err := os.WriteFile(volumeServerIdPath, volumeServerId[:], 0644); err != nil {
			return uuid.Nil, fmt.Errorf("failed to write new volume.id file %s: %v", volumeServerIdPath, err)
		}
		return volumeServerId, nil
	}

	// Handle other errors
	return uuid.Nil, fmt.Errorf("failed reading volume.id file %s: %v", volumeServerIdPath, err)
}
