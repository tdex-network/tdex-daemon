package domain

type Page struct {
	Number int
	Size   int
}

func NewPage(pageNumber, pageSize int) Page {
	pNumber := 1
	if pageNumber > 0 {
		pNumber = pageNumber
	}

	pSize := 10
	if pageSize > 0 {
		pSize = pageSize
	}

	return Page{
		Number: pNumber,
		Size:   pSize,
	}
}
