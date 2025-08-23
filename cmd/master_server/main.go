package main

import (
	"flag"
	"log"

	"github.com/rxanders35/graphene/pkg/master_server"
)

func main() {
	masterAddr := flag.String("master-addr", "localhost:9090", "master's grpc address")

	flag.Parse()
	log.Printf("Starting")

	s := master_server.NewGRPCServer(*masterAddr)
	s.Run()
}
