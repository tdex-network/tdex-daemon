package domain

const (
	DefaultPageNumber = 1
	DefaultPageSize   = 10
)

type Page struct {
	Number int
	Size   int
}

func NewPage(pageNumber, pageSize int) Page {
	pNumber := DefaultPageNumber
	if pageNumber > 0 {
		pNumber = pageNumber
	}

	pSize := DefaultPageSize
	if pageSize > 0 {
		pSize = pageSize
	}

	return Page{
		Number: pNumber,
		Size:   pSize,
	}
}
