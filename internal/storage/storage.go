package storage

import "errors"

var (
	ErrFileNameReplication = errors.New("such file name already taken")
	//ErrFileNotFound        = errors.New("file not found") - not required cause file will be updated
)
