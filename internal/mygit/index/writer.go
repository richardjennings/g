package index

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"github.com/richardjennings/mygit/internal/mygit/config"
	"io"
	"os"
)

// Write writes an Index struct to the Git Index
func (idx *Index) Write() error {
	if idx.header.NumEntries != uint32(len(idx.items)) {
		return errors.New("index numEntries and length of items inconsistent")
	}
	path := config.IndexFilePath()
	f, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	// use a multi-writer to allow both writing the the file whilst incrementally generating
	// a Sha hash of the content as it is written
	h := sha1.New()
	mw := io.MultiWriter(f, h)

	// write header
	if err := binary.Write(mw, binary.BigEndian, idx.header); err != nil {
		return err
	}
	// write each item fixed size entry
	for _, item := range idx.items {
		if err := binary.Write(mw, binary.BigEndian, item.indexItemP); err != nil {
			return err
		}
		// write name
		if _, err := mw.Write(item.Name); err != nil {
			return err
		}
		// write padding
		padding := make([]byte, 8-(62+len(item.Name))%8)
		if _, err := mw.Write(padding); err != nil {
			return err
		}
	}
	// use the generated hash
	sha := h.Sum(nil)
	copy(idx.sig[:], sha)
	// write Sha hash of Index
	if err := binary.Write(f, binary.BigEndian, &sha); err != nil {
		return err
	}

	return f.Close()
}
