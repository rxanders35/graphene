package volume_server

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
)

func TestVolume_WriteRead(t *testing.T) {
	tempDir := t.TempDir()
	volumeID := uuid.New()
	v, err := NewVolume(tempDir, volumeID)
	if err != nil {
		t.Fatalf("Failed to create new volume: %v", err)
	}

	testData := []byte("hello, world")
	objectID, err := v.Write(testData)
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	retrievedData, err := v.Read(objectID)
	if err != nil {
		t.Fatalf("Read() failed: %v", err)
	}

	if !bytes.Equal(testData, retrievedData) {
		t.Fatalf("data mismatch: expected %s, got %s", testData, retrievedData)
	}

	v.dataFile.Close()
	v.idxFile.Close()

	reloadedVolume, err := NewVolume(tempDir, volumeID)
	if err != nil {
		t.Fatalf("Failed to reload volume from disk: %v", err)
	}

	reloadedData, err := reloadedVolume.Read(objectID)
	if err != nil {
		t.Fatalf("Read() from reloaded volume failed: %v", err)
	}

	if !bytes.Equal(testData, reloadedData) {
		t.Fatalf("data mismatch after reload: expected %s, got %s", testData, reloadedData)
	}

	t.Log("Write, Read, and Reload tests passed successfully.")
}
