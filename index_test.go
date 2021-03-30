package fts

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveLoad(t *testing.T) {
	const fileName = "temp1.tmp"
	defer os.Remove(fileName)

	index, err := Open(fileName)
	assert.NoError(t, err)

	index.Add(1, "first document")
	err = index.Save()
	assert.NoError(t, err)
	assert.Equal(t, map[string]int64{"document": 0, "first": 13}, index.filePositions)

	index.Add(2, "second document")
	err = index.Save()
	assert.NoError(t, err)
	assert.Equal(t, map[string]int64{"document": 0, "first": 14, "second": 24}, index.filePositions)

	index2, err := Open(fileName)
	assert.NoError(t, err)
	assert.Len(t, index2.filePositions, 3)
}
