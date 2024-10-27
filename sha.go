package g

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

type (
	Sha struct {
		set  bool
		hash [20]byte
	}
)

// NewSha creates a Sha from either a binary or hex encoded byte slice
func NewSha(b []byte) (Sha, error) {
	if len(b) == 40 {
		s := Sha{set: true}
		_, _ = hex.Decode(s.hash[:], b)
		return s, nil
	}
	if len(b) == 20 {
		s := Sha{set: true}
		copy(s.hash[:], b)
		return s, nil
	}
	return Sha{}, fmt.Errorf("invalid sha %s", b)
}

func ShaFromHexString(s string) (Sha, error) {
	v, err := hex.DecodeString(s)
	if err != nil {
		return Sha{}, err
	}
	return NewSha(v)
}

func (s Sha) Matches(ss Sha) bool {
	return bytes.Compare(s.hash[:], ss.hash[:]) == 0
}

func (s Sha) String() string {
	return s.AsHexString()
}

func (s Sha) IsSet() bool {
	return s.set
}

func (s Sha) AsHexString() string {
	return hex.EncodeToString(s.hash[:])
}

func (s Sha) AsHexBytes() []byte {
	b := make([]byte, 40)
	hex.Encode(b, s.hash[:])
	return b
}

func (s Sha) AsArray() [20]byte {
	var r [20]byte
	copy(r[:], s.hash[:])
	return r
}

// AsByteSlice returns a Sha as a byte slice
func (s Sha) AsByteSlice() []byte {
	return s.hash[:]
}
