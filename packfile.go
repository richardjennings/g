package g

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type (
	PackObjectType uint8
)

const (
	_ PackObjectType = iota
	ObjCommit
	ObjTree
	ObjBlob
	ObjTag
	ObjOfsDelta
	ObjRefDelta
)

// PackFileReadCloser is a Factory that creates a ReadCloser for reading Object
// content from a Pack File that is Not Deltified.
var PackFileReadCloser = func(path string, offset int64) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		fh, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("opening packfile %s: %w", path, err)
		}
		defer func() { _ = fh.Close() }()
		_, err = fh.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("seeking in packfile %s: %w", path, err)
		}
		r, err := zlib.NewReader(fh)
		if err != nil {
			return nil, fmt.Errorf("decompressing packfile %s: %w", path, err)
		}
		return r, nil
	}
}

var PackFileReadCloserRefDelta = func(path string, offset int64) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		return nil, errors.New("not implemented")
	}
}

var PackFileReadCloserOfsDelta = func(path string, offset int64) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		return nil, errors.New("not implemented")
	}
}

func lookupInPackfiles(sha Sha) (*Object, error) {
	var packFiles []string
	// find the available pack files
	if err := filepath.Walk(
		ObjectPackfileDirectory(),
		func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			path = filepath.Base(path)
			if filepath.Ext(info.Name()) == ".idx" {
				packFiles = append(packFiles, path[5:len(path)-4])
			}
			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("walking packfile directory: %w", err)
	}
	// check each pack file index for the sha
	for _, v := range packFiles {
		offset, found, err := findOffsetInIdx(sha, filepath.Join(ObjectPackfileDirectory(), fmt.Sprintf("pack-%s.idx", v)))
		if err != nil {
			return nil, fmt.Errorf("searching pack index %s: %w", v, err)
		}
		if found {
			return findObjectInPack(offset, filepath.Join(ObjectPackfileDirectory(), fmt.Sprintf("pack-%s.pack", v)), sha)
		}
	}
	return nil, nil
}

func readIdxMagic(fh *os.File) error {
	magic := make([]byte, 4)
	if err := binary.Read(fh, binary.BigEndian, magic); err != nil {
		return fmt.Errorf("reading idx magic: %w", err)
	}
	// check magic bytes
	if magic[0] != 255 || magic[1] != 116 || magic[2] != 79 || magic[3] != 99 {
		return errors.New("invalid packfile index magic bytes")
	}
	return nil
}

func readPackMagic(fh *os.File) error {
	magic := make([]byte, 4)
	if err := binary.Read(fh, binary.BigEndian, magic); err != nil {
		return fmt.Errorf("reading pack magic: %w", err)
	}
	// check magic bytes
	if magic[0] != 80 || magic[1] != 65 || magic[2] != 67 || magic[3] != 75 {
		return errors.New("invalid packfile index magic bytes")
	}
	return nil
}

func readIdxFormat(fh *os.File) (uint32, error) {
	var format uint32
	if err := binary.Read(fh, binary.BigEndian, &format); err != nil {
		return 0, fmt.Errorf("reading idx format: %w", err)
	}
	return format, nil
}

func readPackFormat(fh *os.File) (uint32, error) {
	var format uint32
	if err := binary.Read(fh, binary.BigEndian, &format); err != nil {
		return 0, fmt.Errorf("reading pack format: %w", err)
	}
	return format, nil
}

func readFanout(fh *os.File) ([256]uint32, error) {
	// fanout is an array off jump offsets for the first byte of a sha
	// this allows us to search for a sha faster, by starting closer.
	var fanout [256]uint32
	if err := binary.Read(fh, binary.BigEndian, &fanout); err != nil {
		return fanout, fmt.Errorf("reading fanout table: %w", err)
	}
	return fanout, nil
}

func findObjectName(items uint32, fh *os.File, sha Sha) (uint32, bool, error) {
	if items == 0 {
		return 0, false, nil
	}
	// read all SHA hashes in this bucket into memory for binary search
	buf := make([]byte, items*20)
	if _, err := io.ReadFull(fh, buf); err != nil {
		return 0, false, fmt.Errorf("reading object names: %w", err)
	}
	// binary search - pack idx entries are sorted by SHA
	target := sha.hash
	lo, hi := uint32(0), items
	for lo < hi {
		mid := lo + (hi-lo)/2
		h := buf[mid*20 : mid*20+20]
		if bytes.Compare(h, target[:]) < 0 {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo < items && bytes.Equal(buf[lo*20:lo*20+20], target[:]) {
		return lo, true, nil
	}
	return 0, false, nil
}

func readObjectOffset(size uint32, fh *os.File, i uint32) (uint32, error) {
	// skip remaining sorted object names
	// skip 4-byte CRC32 values (*size)
	// skip to i offset in 4 byte offset values
	// @todo if offset most significant bit is set, lookupInPackfiles in long offset table
	if _, err := fh.Seek(int64(4+4+(256*4)+(20*size)+(4*size)+(4*i)), io.SeekStart); err != nil {
		return 0, fmt.Errorf("seeking to object offset: %w", err)
	}
	var offset uint32
	if err := binary.Read(fh, binary.BigEndian, &offset); err != nil {
		return 0, fmt.Errorf("reading object offset: %w", err)
	}
	// we now have the offset to lookupInPackfiles in the pack
	return offset, nil
}

func findOffsetInIdx(sha Sha, path string) (uint32, bool, error) {
	fh, err := os.Open(path)
	if err != nil {
		return 0, false, fmt.Errorf("opening idx %s: %w", path, err)
	}
	defer func() { _ = fh.Close() }()
	// read the magic bytes to check correct
	if err := readIdxMagic(fh); err != nil {
		return 0, false, fmt.Errorf("idx %s: %w", path, err)
	}
	// read the idx format and assert it is 2
	if format, err := readIdxFormat(fh); err != nil || format != 2 {
		if err != nil {
			return 0, false, fmt.Errorf("idx %s: %w", path, err)
		} else {
			return 0, false, fmt.Errorf("invalid pack file idx format in %s, expected 2", path)
		}
	}
	// read fanout buckets
	fanout, err := readFanout(fh)
	if err != nil {
		return 0, false, fmt.Errorf("idx %s: %w", path, err)
	}
	// lookupInPackfiles search bounds
	var startOffset uint32
	if sha.hash[0] == 0 {
		startOffset = 0
	} else {
		startOffset = fanout[sha.hash[0]-1]
	}
	endOffset := fanout[sha.hash[0]]
	size := fanout[255]

	// to make the search more efficient, we can jump to the start
	// address of this sha 1st byte bucket.
	if _, err := fh.Seek(int64(startOffset*20), io.SeekCurrent); err != nil {
		return 0, false, fmt.Errorf("seeking to fanout bucket: %w", err)
	}

	i, found, err := findObjectName(endOffset-startOffset, fh, sha)
	if err != nil {
		return 0, false, fmt.Errorf("searching for object %s: %w", sha, err)
	}
	if !found {
		return 0, false, nil
	}

	offset, err := readObjectOffset(size, fh, i+startOffset)
	if err != nil {
		return 0, false, fmt.Errorf("reading offset for object %s: %w", sha, err)
	}
	return offset, found, nil
}

func findObjectInPack(offset uint32, path string, sha Sha) (*Object, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening packfile %s: %w", path, err)
	}
	defer func() { _ = fh.Close() }()

	if err := readPackMagic(fh); err != nil {
		return nil, fmt.Errorf("packfile %s: %w", path, err)
	}

	// read the idx format and assert it is 2
	if format, err := readPackFormat(fh); err != nil || format != 2 {
		if err != nil {
			return nil, fmt.Errorf("packfile %s: %w", path, err)
		} else {
			return nil, fmt.Errorf("invalid pack file format in %s, expected 2", path)
		}
	}

	var size uint32
	if err := binary.Read(fh, binary.BigEndian, &size); err != nil {
		return nil, fmt.Errorf("reading packfile %s size: %w", path, err)
	}

	if _, err := fh.Seek(int64(offset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("seeking in packfile %s: %w", path, err)
	}

	typ, length, err := readPackTypeLength(fh)
	if err != nil {
		return nil, fmt.Errorf("reading pack object type in %s: %w", path, err)
	}

	obj := &Object{}
	switch typ {
	case ObjBlob, ObjOfsDelta, ObjRefDelta:
		obj.Typ = ObjectTypeBlob
	case ObjCommit, ObjTag:
		obj.Typ = ObjectTypeCommit
	case ObjTree:
		obj.Typ = ObjectTypeTree
	}
	obj.Sha = sha
	obj.Length = int(length)
	// This HeaderLength was added before I knew about pack files,
	// the purpose was to create a factory that allowed a reader to
	// be initialized if required. The HeaderLength bytes are discarded
	// before a ReadCloser implementation streams object content via zlib.
	// For pack files this still makes sense but here we set it to 0 and
	// include the seeking as part of the ReadCloser factory config.
	p, err := fh.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("getting position in packfile %s: %w", path, err)
	}
	obj.HeaderLength = 0
	switch typ {
	case ObjOfsDelta:
		obj.ReadCloser = PackFileReadCloserOfsDelta(path, p)
	case ObjRefDelta:
		obj.ReadCloser = PackFileReadCloserRefDelta(path, p)
	default:
		obj.ReadCloser = PackFileReadCloser(path, p)
	}

	return obj, nil
}

func readPackTypeLength(fh *os.File) (PackObjectType, uint64, error) {
	var v uint8
	var t PackObjectType
	var l uint64
	for i := 0; i < 9; i++ {
		if err := binary.Read(fh, binary.BigEndian, &v); err != nil {
			return 0, 0, fmt.Errorf("reading pack type/length byte: %w", err)
		}
		if i == 0 {
			t = PackObjectType(v & 0b01110000 >> 4)
			l = uint64(v & 0b00001111)
		} else {
			l |= uint64(v&0b01111111) << (4 + ((i - 1) * 7))
		}
		if v&0b10000000 == 0 {
			// no continue bit set
			break
		}
	}
	return t, l, nil
}
