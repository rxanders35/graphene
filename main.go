package main

import (
	"log"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("make run [master || worker <port>]")
	}

	switch os.Args[1] {
	case "master":
		//startMaster()
	case "worker":
		if len(os.Args) < 3 {
			log.Fatal("Not enough args")
		}
		startWorker(os.Args[2])
	}
}
