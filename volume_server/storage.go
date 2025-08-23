package volume_server

import "github.com/google/uuid"

type Storage interface {
	Write(data []byte) (uuid.UUID, error)
	Read(id uuid.UUID) ([]byte, error)
}
