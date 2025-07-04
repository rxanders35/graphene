package volume

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/rxanders35/sss/storage"
)

type Volume struct {
	ID       [16]byte
	idxFile  *os.File
	dataFile *os.File
	idxMap   map[[16]byte]storage.IndexEntry
	rw       sync.RWMutex
}

func NewVolume(path string, volumeID [16]byte) (*Volume, error) {
	fileName := fmt.Sprintf("%s%x", storage.VolumeFilePrefix, volumeID)

	idxFilePath := filepath.Join(path, fileName+storage.IdxFileExtension)
	dataFilePath := filepath.Join(path, fileName+storage.DataFileExtension)

	idxFile, err := os.OpenFile(idxFilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("Failed to instantiate index file for volume server: %v", err)
	}

	dataFile, err := os.OpenFile(dataFilePath, os.O_CREATE|os.O_RDWR, 0666)

	idxMap, err := loadIndex(idxFile)
	if err != nil {
		return nil, err
	}

	return &Volume{
		ID:       volumeID,
		idxFile:  idxFile,
		dataFile: dataFile,
		idxMap:   idxMap,
	}, nil
}

func loadIndex(file *os.File) (map[[16]byte]storage.IndexEntry, error) {
	idxMap := make(map[[16]byte]storage.IndexEntry)

	stats, err := file.Stat()
	if err != nil {
		return nil, err
	}
	idxSize := stats.Size()
	if idxSize%storage.IdxEntryTotalSize != 0 {
		return nil, errors.New("Index file corrupted: not evenly divisible by 28")
	}

	numEntries := int(idxSize) / storage.IdxEntryTotalSize
	entryBuf := make([]byte, storage.IdxEntryTotalSize)

	for i := 0; i < numEntries; i++ {
		_, err := io.ReadFull(file, entryBuf)
		if err != nil {
			return nil, err
		}
		id, entry, err := decodeEntry(entryBuf)
		if err != nil {
			return nil, errors.New("Massive issue decoding index data into its tracker map")
		}

		idxMap[id] = entry

	}
	return idxMap, nil
}

func decodeEntry(buf []byte) ([16]byte, storage.IndexEntry, error) {
	if len(buf) != storage.IdxEntryTotalSize {
		return [16]byte{}, storage.IndexEntry{}, io.ErrUnexpectedEOF
	}
	var id [16]byte
	copy(id[:], buf[0:16])

	var entry storage.IndexEntry
	entry.Offset = binary.BigEndian.Uint64(buf[16:24])
	entry.Size = binary.BigEndian.Uint32(buf[24:28])

	return id, entry, nil
}
