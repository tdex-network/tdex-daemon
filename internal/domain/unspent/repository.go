package unspent

type Repository interface {
	AddUnspent(unspent []Unspent)
	GetAllUnspent() []Unspent
	GetBalance(address string, assetHast string) uint64
	GetAvailableUnspent() []Unspent
}
