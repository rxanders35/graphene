package main

import (
	"log"
	"os"
	"sync"
)

type Worker struct {
	mu         sync.Mutex
	needleFile os.File
	idxFile    os.File

	idx map[[16]byte]struct {
		offset int64
		size   int32
	}
}

func newWorker(port string) (error, *Worker) {
	needleFileName := dataPref + port + dataExt
	needle, err := os.OpenFile(needleFileName, 0666, os.FileMode(os.O_RDWR|os.O_APPEND|os.O_CREATE))
	if err != nil {
		return err, nil
	}

	idxFileName := idxPref + port + idxExt
	idx, err := os.OpenFile(idxFileName, 0666, os.FileMode(os.O_RDWR|os.O_APPEND|os.O_CREATE))
	if err != nil {
		return err, nil
	}

	ifs, err := os.Stat(idxFileName)
	if ifs.Size()%idxSizeTotal != 0 {
		return err, nil
		log.Print("corrupt idx file")
	}

	chunk := ifs.Size() / 28
	for i := int64(0); i < chunk; i++ {
		// TODO
	}

	return nil, &Worker{
		needleFile: *needle,
	}

}
