package components

import (
	"os"
	"sync"

	"github.com/google/uuid"
)

type Worker struct {
	mu         sync.Mutex
	needleFile os.File
	idxFile    os.File

	idx map[uuid.UUID]struct {
		offset int64
		size   int32
	}
}

func newWorker(port string) (*Worker, error) {
}
