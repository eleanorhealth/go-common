package bao

import "errors"

var (
	ErrModelNotStructOrSlice = errors.New("model must be a struct or a slice")
	ErrModelNotStruct        = errors.New("model must be a pointer to a struct")
	ErrOnePrimaryKey         = errors.New("table must have exactly one primary key")
	ErrUpdateNotExists       = errors.New("model to be updated does not exist")
	ErrIDNotUUID             = errors.New("id must be a valid UUID")
)
