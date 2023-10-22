package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

// A file that only writes to disk if the contents on disk
// changed.
// This is useful to avoid re-triggers of fsnotify watchers.
type IfChangedFile struct {
	path string
	buf  *bytes.Buffer
}

func (f *IfChangedFile) Write(p []byte) (int, error) {
	return f.buf.Write(p)
}

func (f *IfChangedFile) Close() error {
	write := true
	_, err := os.Stat(f.path)
	if err == nil {
		ef, err := os.Open(f.path)
		if err == nil {
			defer ef.Close()
			h1 := sha256.New()
			_, err = io.Copy(h1, ef)
			if err == nil {
				hash1 := hex.EncodeToString(h1.Sum(nil))
				h2 := sha256.New()
				if _, err = h2.Write(f.buf.Bytes()); err == nil {
					hash2 := hex.EncodeToString(h2.Sum(nil))
					if hash1 == hash2 {
						write = false
					}
				}
			}
			ef.Close()
		}
	}

	if write {
		if err := os.MkdirAll(filepath.Dir(f.path), 0755); err != nil {
			return err
		}
		osf, err := os.Create(f.path)
		if err != nil {
			return err
		}
		_, err = osf.Write(f.buf.Bytes())
		if err != nil {
			return err
		}
		return osf.Close()
	}
	return nil
}

func CreateIfChangedFile(path string) *IfChangedFile {
	return &IfChangedFile{path: path, buf: &bytes.Buffer{}}
}
