package lsm

import "github.com/cockroachdb/pebble"

type Option func(*LSM)

func WithCache(size int64) Option {
	return func(l *LSM) {
		l.opts.Cache = pebble.NewCache(size)
	}
}

func WithMemTableSize(size uint64) Option {
	return func(l *LSM) {
		l.opts.MemTableSize = size
	}
}

func WithBytesPerSync(bytes int) Option {
	return func(l *LSM) {
		l.opts.BytesPerSync = bytes
	}
}
