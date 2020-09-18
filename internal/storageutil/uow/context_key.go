package uow

const (
	// InMemoryContextKey is the context key that must be shared among multiple
	// reporitories involved in a UnitOfWork transaction
	InMemoryContextKey = iota
)
