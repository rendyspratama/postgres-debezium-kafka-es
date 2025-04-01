package utils

type Pagination struct {
	Limit  int
	Offset int
}

func NewPagination(page, size int) Pagination {
	if size > 100 {
		size = 100
	}
	return Pagination{
		Limit:  size,
		Offset: (page - 1) * size,
	}
}
