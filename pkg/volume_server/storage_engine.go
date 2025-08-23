package volume_server

import "github.com/google/uuid"

type StorageEngine interface {
	Write(id uuid.UUID, data []byte) error
	Read(id uuid.UUID) ([]byte, error)
}
