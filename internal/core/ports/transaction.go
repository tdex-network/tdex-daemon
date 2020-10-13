package ports

type DbManager interface {
	NewTransaction() Transaction
}

type Transaction interface {
	Commit() error
	Discard()
}
