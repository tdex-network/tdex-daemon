package ports

// DbManager interface defines the methods for swap, price and unspent.
type DbManager interface {
	NewTransaction() Transaction
	NewPricesTransaction() Transaction
	NewUnspentsTransaction() Transaction
	IsTransactionConflict(err error) bool
}

// Transaction interface defines the method to commit or discard a database transaction.
type Transaction interface {
	Commit() error
	Discard()
}
