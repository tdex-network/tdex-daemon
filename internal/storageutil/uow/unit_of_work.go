package uow

import (
	"context"
	"fmt"
)

// Transactional begins a transaction
type Transactional interface {
	Begin() (Tx, error)
}

// Tx represents an all-or-nothing transaction, by committing or rolling back
// a set of read/write operations
type Tx interface {
	Commit() error
	Rollback() error
}

// Contextual returns a context for a given argument.
type Contextual interface {
	Context(interface{}) context.Context
}

// ContextProvider returns a context key
type ContextProvider interface {
	ContextKey() interface{}
}

// UnitOfWork allows to run multiple transactions as one
type UnitOfWork struct {
	repositories []Transactional
	contexts     map[interface{}]interface{}
}

// NewUnitOfWork returns a new UnitOfWork with the given Transaction interfaces
func NewUnitOfWork(repositories ...Transactional) *UnitOfWork {
	return &UnitOfWork{
		repositories: repositories,
		contexts:     map[interface{}]interface{}{},
	}
}

// Context returns the context for the given argument.
func (u *UnitOfWork) Context(repository interface{}) context.Context {
	return context.WithValue(context.Background(), repository, u.contexts[repository])
}

// Run executes the given function over the current UnitOfWork. The given
// function is likely making read/write operations to different repositories in
// a transactional way. Run makes sure that all the transactions within the
// given function are either all committed to the relative storage or rolled
// back if any error occur
func (u *UnitOfWork) Run(fn func(Contextual) error) (err error) {
	txs := make([]Tx, 0, len(u.repositories))

	defer func() {
		if err == nil {
			return
		}
		for _, tx := range txs {
			if _err := tx.Rollback(); _err != nil {
				// TODO: handle rollback failure
				// - Shall we try to rollback again?
				err = _err
				return
			}
		}
	}()

	defer func() {
		if err != nil {
			return
		}
		for _, tx := range txs {
			if _err := tx.Commit(); _err != nil {
				// TODO: handle Commit failure.
				// - Shall we try to commit again? If no, how to rollback previously committed transactions?
				err = _err
				return
			}
		}
	}()

	defer func() {
		// panicking returns an error that causes txs rollback
		if rec := recover(); rec != nil {
			err = fmt.Errorf("recovered: %v", rec)
		}
	}()

	for _, r := range u.repositories {
		var key interface{} = r
		if cp, ok := r.(ContextProvider); ok {
			key = cp.ContextKey()
		}
		// make sure that the same context providers share the same context
		if _, ok := u.contexts[key]; ok {
			continue
		}

		tx, err := r.Begin()
		if err != nil {
			return err
		}
		u.contexts[key] = tx
		txs = append(txs, tx)
	}

	return fn(u)
}
