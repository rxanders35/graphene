package volume_server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

var rwrwrw int = 0666

type NeedleVolume struct {
	volumeID [16]byte
	idxFile  *os.File
	dataFile *os.File
	idxMap   map[[16]byte]IndexEntry
	rw       sync.RWMutex
}

func NewNeedleVolume(path string, volumeID [16]byte) (*NeedleVolume, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("could not create data directory: %w", err)
	}

	fileName := fmt.Sprintf("%s%x", VolumeFilePrefix, volumeID)
	idxFilePath := filepath.Join(path, fileName+IdxFileExtension)
	dataFilePath := filepath.Join(path, fileName+DataFileExtension)

	idxFile, err := os.OpenFile(idxFilePath, os.O_CREATE|os.O_RDWR, os.FileMode(rwrwrw))
	if err != nil {
		log.Fatalf("Failed to instantiate index file for volume server: %v", err)
	}

	dataFile, err := os.OpenFile(dataFilePath, os.O_CREATE|os.O_RDWR, os.FileMode(rwrwrw))
	if err != nil {
		log.Fatalf("Failed to instantiate data file for volume server: %v", err)
	}

	idxMap, err := loadIndex(idxFile)
	if err != nil {
		return nil, err
	}

	return &NeedleVolume{
		volumeID: volumeID,
		idxFile:  idxFile,
		dataFile: dataFile,
		idxMap:   idxMap,
	}, nil
}

func (v *NeedleVolume) Write(data []byte) (uuid.UUID, error) {
	v.rw.Lock()
	defer v.rw.Unlock()

	needleId := uuid.New()

	info, err := v.dataFile.Stat()
	if err != nil {
		return needleId, err
	}

	offset := info.Size()
	checksum := crc32.ChecksumIEEE(data)

	needleSize := NeedleFixedPortion + len(data)
	newNeedleBuffer := make([]byte, needleSize)

	binary.BigEndian.PutUint16(newNeedleBuffer[0:2], NeedleMagicVal)
	copy(newNeedleBuffer[2:18], needleId[:])
	binary.BigEndian.PutUint32(newNeedleBuffer[18:22], uint32(len(data)))
	copy(newNeedleBuffer[22:22+len(data)], data)
	binary.BigEndian.PutUint32(newNeedleBuffer[22+len(data):], checksum)

	if _, err := v.dataFile.Write(newNeedleBuffer); err != nil {
		return needleId, err
	}

	idxBuf := make([]byte, IdxEntryTotalSize)
	copy(idxBuf[0:16], needleId[:])
	binary.BigEndian.PutUint64(idxBuf[16:24], uint64(offset))
	binary.BigEndian.PutUint32(idxBuf[24:28], uint32(len(data)))

	if _, err := v.idxFile.Write(idxBuf); err != nil {
		return needleId, err
	}

	v.idxMap[needleId] = IndexEntry{
		Offset: uint64(offset),
		Size:   uint32(len(data)),
	}

	return needleId, nil

}

func (v *NeedleVolume) Read(id uuid.UUID) ([]byte, error) {
	v.rw.RLock()
	defer v.rw.RUnlock()

	entry, ok := v.idxMap[id]
	if !ok {
		return nil, errors.New("catastrophic error: we didn't find the object")
	}

	needleSize := NeedleFixedPortion + int(entry.Size)
	needleBuf := make([]byte, needleSize)

	_, err := v.dataFile.ReadAt(needleBuf, int64(entry.Offset))
	if err != nil {
		return nil, fmt.Errorf("couldnt read needle: %w", err)
	}

	if binary.BigEndian.Uint16(needleBuf[0:2]) != NeedleMagicVal {
		return nil, errors.New("CORRUPTED: even the magic number aint right")
	}

	data := needleBuf[22 : 22+entry.Size]
	onDiskChecksum := binary.BigEndian.Uint32(needleBuf[22+entry.Size:])
	if onDiskChecksum != crc32.ChecksumIEEE(data) {
		return nil, errors.New("CORRUPTED: checksums are totally different")
	}

	return data, nil
}

func loadIndex(file *os.File) (map[[16]byte]IndexEntry, error) {
	idxMap := make(map[[16]byte]IndexEntry)

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	idxSize := info.Size()
	if idxSize%IdxEntryTotalSize != 0 {
		return nil, errors.New("index file corrupted: not evenly divisible by 28")
	}

	numEntries := int(idxSize) / IdxEntryTotalSize
	entryBuf := make([]byte, IdxEntryTotalSize)

	for i := 0; i < numEntries; i++ {
		_, err := io.ReadFull(file, entryBuf)
		if err != nil {
			return nil, err
		}
		id, entry, err := decodeEntry(entryBuf)
		if err != nil {
			return nil, errors.New("massive issue decoding index data into its tracker map fields")
		}

		idxMap[id] = entry

	}
	return idxMap, nil
}

func decodeEntry(buf []byte) ([16]byte, IndexEntry, error) {
	if len(buf) != IdxEntryTotalSize {
		return uuid.UUID{}, IndexEntry{}, io.ErrUnexpectedEOF
	}

	var id uuid.UUID
	copy(id[:], buf[0:16])

	var entry IndexEntry
	entry.Offset = binary.BigEndian.Uint64(buf[16:24])
	entry.Size = binary.BigEndian.Uint32(buf[24:28])

	return id, entry, nil
}
