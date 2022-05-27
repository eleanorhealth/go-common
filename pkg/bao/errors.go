package bao

import "errors"

var ErrModelNotPointer = errors.New("model must be a pointer")
var ErrModelNotStructSlicePointer = errors.New("model must be a struct pointer or slice pointer")
var ErrOnePrimaryKey = errors.New("table must have exactly one primary key")
