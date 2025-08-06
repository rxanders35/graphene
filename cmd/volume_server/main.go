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

	httpSrv, err := volume_server.NewHTTPServer(*volumeHTTPAddr, v, m, serverId)
	if err != nil {
		log.Fatalf("Couldn't init volume server. Why: %v", err)
	}

	go func() {
		if err := httpSrv.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server run error. Why: %v", err)
		}
	}()

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
	volumeServerIdPath := dataDir + "/volume.id"
	idData, err := os.ReadFile(volumeServerIdPath)
	if err == nil {
		volumeServerId, err := uuid.FromBytes(idData)
		if err != nil {
			return uuid.Nil, fmt.Errorf("Failed getting volume server id. Why: %v", err)
		}
		return volumeServerId, nil
	}

	if os.IsNotExist(err) {
		volumeServerId := uuid.New()
		if writeErr := os.WriteFile(volumeServerIdPath, volumeServerId[:], 0644); writeErr != nil {
			return uuid.Nil, fmt.Errorf("could not write new volume.id file: %w", writeErr)
		}
		return volumeServerId, nil
	}
	return uuid.Nil, fmt.Errorf("Failed reading volume.id file. Why: %v", err)
}
