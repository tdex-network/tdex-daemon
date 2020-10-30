package ports

type DbManager interface {
	NewDaemonTransaction() Transaction
	NewUnspentsTransaction() Transaction
}

type Transaction interface {
	Commit() error
	Discard()
}
