package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/rxanders35/graphene/volume_server"
)

func main() {
	mastergRPCAddr := flag.String("master-addr", "localhost:9090", "master's grpc address")
	volumeHTTPAddr := flag.String("addr", ":8080", "volume's http address")
	dataDir := flag.String("data-dir", "./data", "volume's data directory")

	flag.Parse()

	serverId, err := getOrCreateServerID(*dataDir)
	if err != nil {
		log.Fatal(err)
	}

	volume, err := volume_server.NewVolume(*dataDir, serverId)
	if err != nil {
		log.Fatalf("Couldn't init volume backend. Why: %v", err)
	}

	masterClient, err := volume_server.NewMasterClient(*mastergRPCAddr)
	if err != nil {
		log.Fatalf("Couldn't connect to master. Why: %v", err)
	}

	httpSrv, err := volume_server.NewHTTPServer(*volumeHTTPAddr, volume, masterClient, serverId)
	if err != nil {
		log.Fatalf("Couldn't init volume server. Why: %v", err)
	}

	go func() {
		if err := httpSrv.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server run error. Why: %v", err)
		}
	}()

	log.Printf("Volume Server is running on %s", *volumeHTTPAddr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shut down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed. Why: %v", err)
	}
}

func getOrCreateServerID(dataDir string) (uuid.UUID, error) {
	idPath := filepath.Join(dataDir, "volume.id")

	idBytes, err := os.ReadFile(idPath)
	if err == nil {
		id, parseErr := uuid.FromBytes(idBytes)
		if parseErr != nil {
			return uuid.Nil, fmt.Errorf("corrupt volume.id file: %w", parseErr)
		}
		return id, nil
	}

	if os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return uuid.Nil, fmt.Errorf("failed to create data directory: %w", err)
		}
		newID := uuid.New()
		if writeErr := os.WriteFile(idPath, newID[:], 0644); writeErr != nil {
			return uuid.Nil, fmt.Errorf("could not write new volume.id file: %w", writeErr)
		}
		return newID, nil
	}

	return uuid.Nil, fmt.Errorf("could not read volume.id file: %w", err)
}
