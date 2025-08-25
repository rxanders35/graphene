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

const (
	_128MiB = 128 << 20
	_64MiB  = 64 << 20
	_1MiB   = 1 << 20
)

func NewLSM(path string, opts ...Option) *LSM {
	defaultOpts := &pebble.Options{
		Cache:        pebble.NewCache(_128MiB),
		MemTableSize: _64MiB,
		BytesPerSync: _1MiB,
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
