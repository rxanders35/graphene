package components

import (
	"os"
	"sync"

	"github.com/google/uuid"
)

type Worker struct {
	needle os.File
	idx    os.File

	loc map[uuid.UUID]struct {
		offset int64
		size   int32
	}

	mu sync.Mutex
}
