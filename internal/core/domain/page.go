package domain

type Page interface {
	GetNumber() int64
	GetSize() int64
}
