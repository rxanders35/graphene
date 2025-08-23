package lsm

import (
	"log"

	"github.com/cockroachdb/pebble"
	"github.com/google/uuid"
)

type LSM struct {
	db   *pebble.DB
	opts *pebble.Options
}

func NewLSM(path string, opts ...Option) *LSM {
	defaultOpts := &pebble.Options{
		Cache:        pebble.NewCache(128 << 20),
		MemTableSize: 64 << 20,
		BytesPerSync: 1 << 20,
	}

	l := &LSM{
		opts: defaultOpts,
	}

	for _, opt := range opts {
		opt(l)
	}

	db, err := pebble.Open(path, l.opts)
	if err != nil {
		log.Fatalf("Failed to initialize Pebble LSM. Why: %v", err)
	}

	l.db = db
	return l
}

func (l *LSM) Write(data []byte) error {
	u := uuid.New()

	err := l.db.Set(u[:], data, pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

func (l *LSM) Read(id uuid.UUID) ([]byte, error) {
	d, cl, err := l.db.Get(id[:])
	if err != nil {
		return nil, err
	}
	defer cl.Close()

	data := make([]byte, len(d))
	copy(data, d)

	return data, nil
}
