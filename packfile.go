package g

import (
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

// PackFileReadCloser is a Factory that creates a ReadCloser for
// reading Object content from a Pack File that is Not Deltified.
var PackFileReadCloser = func(path string, offset int64) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		fh, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer func() { _ = fh.Close() }()
		_, err = fh.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, err
		}
		return zlib.NewReader(fh)
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
		return nil, err
	}
	// check each pack file index for the sha
	for _, v := range packFiles {
		offset, found, err := findOffsetInIdx(sha, filepath.Join(ObjectPackfileDirectory(), fmt.Sprintf("pack-%s.idx", v)))
		if err != nil {
			return nil, err
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
		return err
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
		return err
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
		return 0, err
	}
	return format, nil
}

func readPackFormat(fh *os.File) (uint32, error) {
	var format uint32
	if err := binary.Read(fh, binary.BigEndian, &format); err != nil {
		return 0, err
	}
	return format, nil
}

func readFanout(fh *os.File) ([256]uint32, error) {
	// fanout is an array off jump offsets for the first byte of a sha
	// this allows us to search for a sha faster, by starting closer.
	var fanout [256]uint32
	err := binary.Read(fh, binary.BigEndian, &fanout)
	return fanout, err
}

func findObjectName(items uint32, fh *os.File, sha Sha) (uint32, bool, error) {
	var hash [20]byte
	// should be an efficiently implemented binary search,
	// for now a blunt force string trauma
	for i := uint32(0); i < items; i++ {
		if err := binary.Read(fh, binary.BigEndian, &hash); err != nil {
			return i, false, err
		}
		h, err := NewSha(hash[:])
		if err != nil {
			return i, false, err
		}
		if h.AsHexString() == sha.AsHexString() {
			return i, true, nil
		}
	}
	return 0, false, nil
}

func readObjectOffset(size uint32, fh *os.File, i uint32) (uint32, error) {

	// skip remaining sorted object names
	// skip 4-byte CRC32 values (*size)
	// skip to i offset in 4 byte offset values
	// @todo if offset most significant bit is set, lookupInPackfiles in long offset table
	if _, err := fh.Seek(int64(4+4+(256*4)+(20*size)+(4*size)+(4*i)), io.SeekStart); err != nil {
		return 0, err
	}
	var offset uint32
	if err := binary.Read(fh, binary.BigEndian, &offset); err != nil {
		return 0, err
	}
	// we now have the offset to lookupInPackfiles in the pack
	return offset, nil
}

func findOffsetInIdx(sha Sha, path string) (uint32, bool, error) {
	fh, err := os.Open(path)
	if err != nil {
		return 0, false, err
	}
	defer func() { _ = fh.Close() }()

	// read the magic bytes to check correct
	if err := readIdxMagic(fh); err != nil {
		return 0, false, err
	}

	// read the idx format and assert it is 2
	if format, err := readIdxFormat(fh); err != nil || format != 2 {
		if err != nil {
			return 0, false, err
		} else {
			return 0, false, errors.New("invalid pack file idx format, expected 2")
		}
	}

	// read fanout buckets
	fanout, err := readFanout(fh)
	if err != nil {
		return 0, false, err
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
		return 0, false, err
	}

	i, found, err := findObjectName(endOffset-startOffset, fh, sha)
	if err != nil {
		return 0, false, err
	}
	if !found {
		return 0, false, nil
	}

	offset, err := readObjectOffset(size, fh, i+startOffset)
	return offset, found, err
}

func findObjectInPack(offset uint32, path string, sha Sha) (*Object, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = fh.Close() }()

	if err := readPackMagic(fh); err != nil {
		return nil, err
	}

	// read the idx format and assert it is 2
	if format, err := readPackFormat(fh); err != nil || format != 2 {
		if err != nil {
			return nil, err
		} else {
			return nil, errors.New("invalid pack file format, expected 2")
		}
	}

	var size uint32
	if err := binary.Read(fh, binary.BigEndian, &size); err != nil {
		return nil, err
	}

	if _, err := fh.Seek(int64(offset), io.SeekStart); err != nil {
		return nil, err
	}

	typ, length, err := readPackTypeLength(fh)
	if err != nil {
		return nil, err
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
		return nil, err
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
			return 0, 0, err
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
