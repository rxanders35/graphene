package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
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

	needleFileName, idxFileName := makeFileName(port)

	needle, err := os.OpenFile(needleFileName, 0666, os.FileMode(os.O_RDWR|os.O_APPEND|os.O_CREATE))
	if err != nil {
		return nil, err
	}
	w.needleFile = needle

	idxFile, err := os.OpenFile(idxFileName, 0666, os.FileMode(os.O_RDWR|os.O_APPEND|os.O_CREATE))
	if err != nil {
		needle.Close()
		return nil, err
	}
	w.idxFile = idxFile

	indexInfo, err := w.idxFile.Stat()
	if err != nil {
		w.Close()
		return nil, err
	}

	if indexInfo.Size()%idxSizeTotal != 0 {
		log.Print("corrupt idx file")
		w.Close()
		return nil, err
	}

	numEntries := indexInfo.Size() / idxSizeTotal
	for i := int64(0); i < numEntries; i++ {
		_, err := idxFile.Seek(i*idxSizeTotal, 0)
		if err != nil {
			w.Close()
			return nil, fmt.Errorf("failed to seek to entry %d: %w", i, err)
		}
		id := [16]byte{}
		n, err := idxFile.Read(id[:])

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

func (w *Worker) Store(id [16]byte, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	needleFileStat, err := w.needleFile.Stat()
	if err != nil {
		return err
	}

	offset := needleFileStat.Size()
	checksum := crc32.ChecksumIEEE(data)

	needleSize := needleMagicSize + needleIDSize + needleDataSize + len(data) + needleChecksum
	needle := make([]byte, needleSize)

	binary.BigEndian.PutUint16(needle[0:2], needleMagicVal)

	copy(needle[2:18], id[:])
	binary.BigEndian.PutUint32(needle[18:22], uint32(len(data)))

	copy(needle[22:22+len(data)], data)
	binary.BigEndian.PutUint32(needle[22+len(data):], checksum)

	n, err := w.needleFile.Write(needle)
	if err != nil {
		return fmt.Errorf("failed to write needle: %w", err)
	}
	if n != needleSize {
		return errIncompleteWrite
	}

	idxEntry := make([]byte, 28)
	copy(idxEntry[0:16], id[:])
	binary.BigEndian.PutUint64(idxEntry[16:24], uint64(offset))
	binary.BigEndian.PutUint32(idxEntry[24:28], uint32(len(data)))
	_, err = w.idxFile.Write(idxEntry)
	if err != nil {
		return err
	}

	w.idx[id] = struct {
		offset int64
		size   int32
	}{offset, int32(len(data))}

	return nil
}

func (w *Worker) Get(id [16]byte) ([]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	entry, exists := w.idx[id]
	if !exists {
		return nil, errObjectNotFound
	}
	offset := entry.offset
	size := entry.size

	_, err := w.needleFile.Seek(offset, 0)
	if err != nil {
		return nil, fmt.Errorf("Failed to seek: %w", err)
	}

	magicBytes := make([]byte, 2)
	n, err := w.needleFile.Read(magicBytes[:])
	if err != nil || n != 2 {
		return nil, fmt.Errorf("failed to read magic number: %w", err)
	}
	if binary.BigEndian.Uint16(magicBytes[:]) != needleMagicVal {
		return nil, fmt.Errorf("invalid needle magic number")
	}

	idBytes := make([]byte, 16)
	n, err = w.needleFile.Read(idBytes[:])
	if err != nil || n != 16 {
		return nil, fmt.Errorf("failed to read Needle ID: %w", err)
	}

	sizeBytes := make([]byte, 4)
	n, err = w.needleFile.Read(sizeBytes[:])
	if err != nil || n != 4 {
		return nil, fmt.Errorf("failed to read data size: %w", err)
	}
	s := binary.BigEndian.Uint32(sizeBytes[:])
	if int32(s) != size {
		return nil, errors.New("needle size mismatch")
	}

	data := make([]byte, entry.size)
	if n, err = w.needleFile.Read(data); err != nil || n != int(entry.size) {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	checksumBytes := make([]byte, 4)
	n, err = w.needleFile.Read(checksumBytes[:])
	if err != nil || n != 4 {
		return nil, fmt.Errorf("failed to read checksum: %w", err)
	}
	storedChecksum := binary.BigEndian.Uint32(checksumBytes[:])
	checksum := crc32.ChecksumIEEE(data)
	if checksum != storedChecksum {
		return nil, errors.New("checksum mismatch")
	}

	return data, nil
}

func (w *Worker) Close() error {
	var errs []error

	if w.needleFile != nil {
		if err := w.needleFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if w.idxFile != nil {
		if err := w.needleFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

var errIncompleteWrite = errors.New("incomplete write to needle file")
var errObjectNotFound = errors.New("object not found in needle file")
