package types

type Pagination[T any] struct {
	Items       []T
	TotalPages  int64
	CurrentPage int64
	PageSize    int64
	TotalItems  int64
}
