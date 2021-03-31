package fts

import (
	"errors"
)

var (
	errMemOnlyIndex = errors.New("memory index can't be saved")
)
