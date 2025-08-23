package ledger

import "github.com/jmoiron/sqlx"

type Ledger struct {
	store Store
}

func NewLedger() *Ledger {
	l := &Ledger{}
	return l
}

type Store interface{}

type PGStore struct {
	db *sqlx.DB
}

func NewPGStore() *PGStore {
	s := &PGStore{}
	return s
}
