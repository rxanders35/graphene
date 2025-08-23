package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rxanders35/graphene/pkg/gateway"
)

func main() {
	masterAddr := flag.String("master-addr", "localhost:9090", "master's grpc address")
	gatewayAddr := flag.String("gateway-addr", "127.0.0.1:8081", "gateway's http address")

	flag.Parse()

	m, err := gateway.NewMasterclient(*masterAddr)
	if err != nil {
		log.Fatalf("Failed to init master client on API gateway. Why: %v", err)
	}

	h, err := gateway.NewGatewayHandler(m)
	s, err := gateway.NewGatewayServer(*gatewayAddr, h)
	if err != nil {
		log.Fatalf("Failed to init API gateway. Why: %v", err)
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
