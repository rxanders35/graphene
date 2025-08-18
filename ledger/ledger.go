package ledger

import "github.com/jmoiron/sqlx"

type Ledger struct {
	store Store
}

func NewLedger() *Ledger {
	l := &Ledger{}
	return l
}

type Store struct {
	db *sqlx.DB
}

func NewStore() *Store {
	s := &Store{}
	return s
}
