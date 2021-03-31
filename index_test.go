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

func TestSearchFromMemory(t *testing.T) {
	index, err := Open("")
	assert.NoError(t, err)

	index.Add(1, "first document")
	index.Add(2, "second document")
	index.Add(3, "third document")

	expectedMap := map[string]map[int]struct{}{
		"first":    map[int]struct{}{1: struct{}{}},
		"second":   map[int]struct{}{2: struct{}{}},
		"third":    map[int]struct{}{3: struct{}{}},
		"document": map[int]struct{}{1: struct{}{}, 2: struct{}{}, 3: struct{}{}},
	}
	assert.Equal(t, expectedMap, index.memoryData)

	ids, err := index.Search("first")
	assert.NoError(t, err)
	assert.Equal(t, []int{1}, ids)

	ids, err = index.Search("second")
	assert.NoError(t, err)
	assert.Equal(t, []int{2}, ids)

	ids, err = index.Search("third")
	assert.NoError(t, err)
	assert.Equal(t, []int{3}, ids)

	ids, err = index.Search("document")
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, ids)

	ids, err = index.Search("document first")
	assert.NoError(t, err)
	assert.Equal(t, []int{1}, ids)
}

func TestSearchFromDisk(t *testing.T) {
	const fileName = "testSearchFromDisk.tmp"
	defer os.Remove(fileName)

	index, err := Open(fileName)
	assert.NoError(t, err)

	index.Add(1, "first document")
	index.Add(2, "second document")
	index.Add(3, "third document")

	err = index.Save()
	assert.NoError(t, err)

	expectedPositions := map[string]int64{"document": 0, "first": 15, "second": 25, "third": 36}
	assert.Equal(t, expectedPositions, index.filePositions)

	ids, err := index.Search("first")
	assert.NoError(t, err)
	assert.Equal(t, []int{1}, ids)

	ids, err = index.Search("second")
	assert.NoError(t, err)
	assert.Equal(t, []int{2}, ids)

	ids, err = index.Search("third")
	assert.NoError(t, err)
	assert.Equal(t, []int{3}, ids)

	ids, err = index.Search("document")
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, ids)

	ids, err = index.Search("document first")
	assert.NoError(t, err)
	assert.Equal(t, []int{1}, ids)
}
