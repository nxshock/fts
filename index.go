package fts

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/gxba/binary"
)

// Index holds basic indexes of documents.
type Index struct {
	fileName      string
	memoryData    map[string]map[int]struct{}
	filePositions map[string]int64

	unsaved bool

	mu sync.RWMutex
}

// Opens creates new index or read index from disk.
func Open(filePath string) (*Index, error) {
	index := &Index{
		fileName:      filePath,
		memoryData:    make(map[string]map[int]struct{}),
		filePositions: make(map[string]int64)}

	if filePath != "" {
		exists, err := isFileExists(filePath)
		if err != nil {
			return nil, err
		}
		if exists {
			err := index.readDiskStats()
			if err != nil {
				return nil, err
			}
		}
	}

	return index, nil
}

func (index *Index) getKey(key string) ([]int, error) {
	memIds, exists := index.memoryData[key]
	if !exists {
		memIds = make(map[int]struct{})
	}

	diskIds, err := index.loadKeyFromDisk(key)
	if err != nil {
		return nil, err
	}

	ids := make(map[int]struct{})

	for _, id := range diskIds {
		ids[id] = struct{}{}
	}

	for id := range memIds {
		ids[id] = struct{}{}
	}

	return mapToArr(ids), nil
}

func (index *Index) loadKeyFromDisk(key string) ([]int, error) {
	pos, exists := index.filePositions[key]
	if !exists {
		return nil, nil
	}

	f, err := os.Open(index.fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = f.Seek(pos, io.SeekStart)
	if err != nil {
		return nil, err
	}

	diskKey, ids, err := read(f)
	if err != nil {
		return nil, err
	}

	if diskKey != key {
		return nil, fmt.Errorf(`expected "%s" key, got "%s"`, key, diskKey)
	}

	return ids, nil
}

// Add adds data to index.
func (index *Index) Add(id int, text string) {
	index.mu.Lock()
	defer index.mu.Unlock()

	for _, token := range analyze(text) {
		if _, exists := index.memoryData[token]; !exists {
			index.memoryData[token] = make(map[int]struct{})
		}
		index.memoryData[token][id] = struct{}{}
	}

	index.unsaved = true
}

// Search queries index for the given text.
func (index *Index) Search(text string) ([]int, error) {
	index.mu.RLock()
	defer index.mu.RUnlock()

	var r []int
	for _, token := range analyze(text) {
		ids, err := index.getKey(token)
		if err != nil {
			return nil, err
		}

		if len(ids) > 0 {
			if r == nil {
				r = ids
			} else {
				r = intersection(r, ids)
			}
		} else {
			// Token doesn't exist.
			return nil, nil
		}
	}
	return r, nil
}

// Save saves index to disk.
func (index *Index) Save() error {
	index.mu.Lock()
	defer index.mu.Unlock()

	if index.fileName == "" {
		return errors.New("memory index can't be saved")
	}

	if !index.unsaved {
		return nil
	}

	var tempFileName = index.fileName + ".tmp"

	f, err := os.Create(tempFileName)
	if err != nil {
		return err
	}
	buf := bufio.NewWriter(f)

	keys := make(map[string]struct{})
	for k := range index.memoryData {
		keys[k] = struct{}{}
	}
	for k := range index.filePositions {
		keys[k] = struct{}{}
	}

	keysArr := make([]string, len(keys))
	i := -1
	for k := range keys {
		i++
		keysArr[i] = k
	}
	sort.Strings(keysArr)

	newFilePositions := make(map[string]int64)

	for _, k := range keysArr {
		ids, err := index.getKey(k)
		if err != nil {
			return err
		}

		pos, _ := f.Seek(0, io.SeekCurrent)
		pos += int64(buf.Buffered())
		newFilePositions[k] = pos

		err = write(buf, k, ids)
		if err != nil {
			return err
		}
	}

	err = buf.Flush()
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	err = os.Rename(tempFileName, index.fileName)
	if err != nil {
		os.Remove(tempFileName)
		return err
	}

	index.memoryData = make(map[string]map[int]struct{})
	index.filePositions = newFilePositions
	index.unsaved = false

	return nil
}

func (index *Index) readDiskStats() error {
	f, err := os.Open(index.fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := bufio.NewReader(f)

	for {
		pos, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		pos -= int64(buf.Buffered())

		word, _, err := read(buf)
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		index.filePositions[word] = pos
	}

	return nil
}

func read(r io.Reader) (key string, ids []int, err error) {
	var keyLen int
	err = binary.Read(r, binary.LittleEndian, &keyLen)
	if err == io.ErrUnexpectedEOF {
		err = io.EOF
	}
	if err != nil {
		return "", nil, err
	}
	if keyLen <= 0 {
		return "", nil, fmt.Errorf("negative len %v", keyLen)
	}

	b := make([]byte, keyLen)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return "", nil, err
	}

	err = binary.Read(bytes.NewReader(b), binary.LittleEndian, &key)
	if err != nil {
		return "", nil, err
	}

	var idsLen int
	err = binary.Read(r, binary.LittleEndian, &idsLen)
	if err != nil {
		return "", nil, err
	}
	if keyLen <= 0 {
		return "", nil, fmt.Errorf("negative len %v", idsLen)
	}

	b = make([]byte, idsLen)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return "", nil, err
	}

	err = binary.Read(bytes.NewReader(b), binary.LittleEndian, &ids)
	if err != nil {
		return "", nil, err
	}

	return key, ids, nil
}

func write(w io.Writer, key string, ids []int) error {
	b := new(bytes.Buffer)

	binary.Write(b, binary.LittleEndian, binary.Size(key))
	binary.Write(b, binary.LittleEndian, key)

	binary.Write(b, binary.LittleEndian, binary.Size(ids))
	binary.Write(b, binary.LittleEndian, ids)

	_, err := b.WriteTo(w)

	return err
}
