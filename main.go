package main

import (
	"flag"
	"log"

	"github.com/google/uuid"
	"github.com/rxanders35/sss/master"
	"github.com/rxanders35/sss/volume"
)

func main() {
	port := flag.String("port", ":8080", "The port for the volume server to listen on.")
	dataDir := flag.String("data-dir", "./data", "The directory to store volume data files.")
	flag.Parse()

	volumeID := uuid.New()
	v, err := volume.NewVolume(*dataDir, volumeID)
	if err != nil {
		log.Fatalf("Failed to initialize volume: %v", err)
	}

	srv := volume.NewServer(*port, v)
	m := master.NewMaster()

	log.Printf("Starting volume server on port %s", *port)
	srv.Start()

	log.Printf("Starting master server on port %s", *port)
	log.Fatal(m.Start(":9000"))
}
