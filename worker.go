package main

import (
	"encoding/binary"
	"log"
	"os"
	"sync"
)

type Worker struct {
	mu         sync.Mutex
	needleFile *os.File
	idxFile    *os.File

	idx map[[16]byte]struct {
		offset int64
		size   int32
	}
}

func newWorker(port string) (*Worker, error) {
	w := &Worker{}
	w.idx = make(map[[16]byte]struct {
		offset int64
		size   int32
	})

	needleFileName := dataPref + port + dataExt

	needle, err := os.OpenFile(needleFileName, 0666, os.FileMode(os.O_RDWR|os.O_APPEND|os.O_CREATE))
	if err != nil {
		return nil, err
	}

	w.needleFile = needle

	idxFileName := idxPref + port + idxExt

	idxFile, err := os.OpenFile(idxFileName, 0666, os.FileMode(os.O_RDWR|os.O_APPEND|os.O_CREATE))
	if err != nil {
		needle.Close()
		return nil, err
	}

	w.idxFile = idxFile

	ifs, err := w.idxFile.Stat()
	if err != nil {
		w.Close()
		return nil, err
	}

	if ifs.Size()%idxSizeTotal != 0 {
		log.Print("corrupt idx file")
		w.Close()
		return nil, err
	}

	var n int

	numEntries := ifs.Size() / idxSizeTotal
	for i := int64(0); i < numEntries; i++ {
		var id [16]byte
		n, err = idxFile.Read(id[:])
		if err != nil || n != 16 {
			w.Close()
			log.Print("failed to read id")
			return nil, err
		}

		offsetBytes := make([]byte, 8)
		n, err = idxFile.Read(offsetBytes)
		if err != nil || n != 8 {
			w.Close()
			log.Print("offset not found")
			return nil, err
		}

		offset := int64(binary.BigEndian.Uint64(offsetBytes))

		sizeBytes := make([]byte, 4)
		n, err = idxFile.Read(sizeBytes)
		if err != nil || n != 4 {
			w.Close()
			log.Print("failed to read size")
			return nil, err
		}

		size := int32(binary.BigEndian.Uint32(sizeBytes))

		w.idx[id] = struct {
			offset int64
			size   int32
		}{offset, size}

	}

	return w, nil
}

func (w *Worker) Close() error {
	var err1, err2 error

	if w.needleFile != nil {
		err1 = w.needleFile.Close()
	}

	if w.idxFile != nil {
		err2 = w.idxFile.Close()
	}

	if err1 != nil {
		return err1
	}

	return err2
}
