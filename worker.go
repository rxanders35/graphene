package main

import (
	"encoding/binary"
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
	w := &Worker{}
	w.idx = make(map[[16]byte]struct {
		offset int64
		size   int32
	})

	needleFileName := dataPref + port + dataExt

	needle, err := os.OpenFile(needleFileName, 0666, os.FileMode(os.O_RDWR|os.O_APPEND|os.O_CREATE))
	if err != nil {
		return err, nil
	}

	idxFileName := idxPref + port + idxExt

	idxFile, err := os.OpenFile(idxFileName, 0666, os.FileMode(os.O_RDWR|os.O_APPEND|os.O_CREATE))
	if err != nil {
		return err, nil
	}

	ifs, err := os.Stat(idxFileName)

	if ifs.Size()%idxSizeTotal != 0 {
		log.Print("corrupt idx file")
		return err, nil
	}

	var n int
	var err error

	numEntries := ifs.Size() / idxSizeTotal
	for i := int64(0); i < numEntries; i++ {
		var id [16]byte
		n, err = idxFile.Read(id[:])
		if err != nil || n != 16 {
			log.Print("failed to read id")
			return err, nil
		}

		offsetBytes := make([]byte, 8)
		n, err = idxFile.Read(offsetBytes)
		if err != nil || n != 8 {
			log.Print("offset not found")
			return err, nil
		}

		offset := int64(binary.BigEndian.Uint64(offsetBytes))

		sizeBytes := make([]byte, 4)
		n, err = idxFile.Read(sizeBytes)
		if err != nil || n != 4 {
			log.Print("failed to read size")
			return err, nil
		}

		size := int32(binary.BigEndian.Uint32(sizeBytes))

		w.idx[id] = struct {
			offset int64
			size   int32
		}{offset, size}

	}

	return nil, &Worker{
		needleFile: *needle,
	}
}
