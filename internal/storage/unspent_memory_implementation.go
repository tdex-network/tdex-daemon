package storage

import (
	"github.com/tdex-network/tdex-daemon/internal/domain/unspent"
	"sync"
)

// InMemoryUnspentRepository represents an in memory storage
type InMemoryUnspentRepository struct {
	unspents []unspent.Unspent
	lock     *sync.RWMutex
}

//NewInMemoryUnspentRepository returns a new empty InMemoryMarketRepository
func NewInMemoryUnspentRepository() *InMemoryUnspentRepository {
	return &InMemoryUnspentRepository{
		unspents: make([]unspent.Unspent, 0),
		lock:     &sync.RWMutex{},
	}
}

func (i *InMemoryUnspentRepository) AddUnspent(unspents []unspent.Unspent) {

	//add new unspent
	for _, newUnspent := range unspents {
		exist := false
		for _, oldUnspent := range i.unspents {
			if newUnspent.Txid == oldUnspent.Txid &&
				newUnspent.Vout == oldUnspent.Vout {
				exist = true
			}
		}
		if !exist {
			i.unspents = append(i.unspents, newUnspent)
		}
	}

	//update spent
	for index, oldUnspent := range i.unspents {
		exist := false
		for _, newUnspent := range unspents {
			if newUnspent.Txid == oldUnspent.Txid &&
				newUnspent.Vout == oldUnspent.Vout {
				exist = true
			}
		}
		if !exist {
			oldUnspent.Spent = true
			i.unspents[index] = oldUnspent
		}
	}
}

func (i *InMemoryUnspentRepository) GetAllUnspent() []unspent.Unspent {
	unspents := make([]unspent.Unspent, 0)
	for _, u := range i.unspents {
		unspents = append(unspents, u)
	}
	return unspents
}

func (i *InMemoryUnspentRepository) GetBalance(
	address string,
	assetHast string,
) uint64 {
	var balance uint64

	for _, u := range i.unspents {
		if u.Address == address && u.AssetHash == assetHast && !u.Spent {
			balance += u.Value
		}
	}

	return balance
}
