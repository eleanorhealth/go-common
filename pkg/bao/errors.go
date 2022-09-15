package bao

import "errors"

var ErrModelNotStructOrSlice = errors.New("model must be a struct or a slice")
var ErrModelNotStruct = errors.New("model must be a pointer to a struct")
var ErrOnePrimaryKey = errors.New("table must have exactly one primary key")
