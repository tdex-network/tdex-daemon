package ports

type DbManager interface {
	NewTransaction() Transaction
	NewUnspentsTransaction() Transaction
}

type Transaction interface {
	Commit() error
	Discard()
}
