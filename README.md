# fts

Basic full text search for Go.

## Usage

```go
// Create new index or open existing from file
index, err := Open("fileName.bin") // specify "" parameter if you need only memory index

// Add data to index.
index.Add(1, "first document text")
index.Add(2, "second document text")
index.Add(3, "third document text")

// Execute query
ids, err := index.Search("first") // ids holds IDs of documents

// Save index to disk
err = index.Save()
```
