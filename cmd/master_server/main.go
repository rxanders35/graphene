package main

import (
	"flag"
	"log"

	"github.com/rxanders35/graphene/pkg/cluster_manager"
)

func main() {
	masterAddr := flag.String("master-addr", "localhost:9090", "master's grpc address")

	flag.Parse()
	log.Printf("Starting")

	s := cluster_manager.NewGRPCServer(*masterAddr)
	s.Run()
}
