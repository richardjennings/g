package g

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// 1. findObjectName -- binary search in packfile index
// ---------------------------------------------------------------------------

// makeSortedHashFile creates a temp file containing n sorted 20-byte hashes.
// It returns the file (seeked to start), the hash at position mid, and a
// hash that is guaranteed not to be present.
func makeSortedHashFile(b *testing.B, n int) (fh *os.File, midSha Sha, missingSha Sha) {
	b.Helper()

	hashes := make([][20]byte, n)
	for i := 0; i < n; i++ {
		h := sha1.Sum([]byte(fmt.Sprintf("hash-entry-%d", i)))
		hashes[i] = h
	}
	sort.Slice(hashes, func(i, j int) bool {
		for k := 0; k < 20; k++ {
			if hashes[i][k] != hashes[j][k] {
				return hashes[i][k] < hashes[j][k]
			}
		}
		return false
	})

	f, err := os.CreateTemp("", "bench-idx-*")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	})

	for _, h := range hashes {
		if _, err := f.Write(h[:]); err != nil {
			b.Fatal(err)
		}
	}

	mid := n / 2
	midSha = Sha{set: true}
	copy(midSha.hash[:], hashes[mid][:])

	// Build a hash that will never appear: all 0xff bytes.
	missingSha = Sha{set: true}
	for i := range missingSha.hash {
		missingSha.hash[i] = 0xff
	}

	if _, err := f.Seek(0, 0); err != nil {
		b.Fatal(err)
	}
	return f, midSha, missingSha
}

func BenchmarkFindObjectName(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		fh, midSha, missingSha := makeSortedHashFile(b, size)

		b.Run(fmt.Sprintf("hit/%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := fh.Seek(0, 0); err != nil {
					b.Fatal(err)
				}
				_, found, err := findObjectName(uint32(size), fh, midSha)
				if err != nil {
					b.Fatal(err)
				}
				if !found {
					b.Fatal("expected to find sha")
				}
			}
		})

		b.Run(fmt.Sprintf("miss/%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := fh.Seek(0, 0); err != nil {
					b.Fatal(err)
				}
				_, found, err := findObjectName(uint32(size), fh, missingSha)
				if err != nil {
					b.Fatal(err)
				}
				if found {
					b.Fatal("expected not to find sha")
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 2. NewSha -- binary (20-byte) and hex (40-byte) paths
// ---------------------------------------------------------------------------

func BenchmarkNewSha(b *testing.B) {
	binInput := sha1.Sum([]byte("bench-sha-input"))

	hexSha, err := NewSha(binInput[:])
	if err != nil {
		b.Fatal(err)
	}
	hexInput := []byte(hexSha.AsHexString())

	b.Run("binary20", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := NewSha(binInput[:]); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("hex40", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := NewSha(hexInput); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 3. ObjectTree -- build a tree from a flat file list
// ---------------------------------------------------------------------------

func BenchmarkObjectTree(b *testing.B) {
	buildFiles := func(n int) []*FileStatus {
		files := make([]*FileStatus, 0, n)
		dummySha := Sha{set: true} // zero-valued hash is fine for structure benchmark
		for i := 0; i < n; i++ {
			var p string
			switch {
			case i%5 == 0:
				p = fmt.Sprintf("dir%d/subdir/file%d.go", i%3, i)
			case i%3 == 0:
				p = fmt.Sprintf("dir%d/file%d.go", i%4, i)
			default:
				p = fmt.Sprintf("file%d.go", i)
			}
			files = append(files, &FileStatus{
				path:  p,
				index: &fileInfo{Sha: dummySha},
			})
		}
		return files
	}

	for _, n := range []int{10, 100} {
		files := buildFiles(n)
		b.Run(fmt.Sprintf("files_%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ObjectTree(files, "")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 4. ReadIndex / Index.Write round-trip
// ---------------------------------------------------------------------------

func BenchmarkReadIndex(b *testing.B) {
	for _, n := range []int{10, 100} {
		b.Run(fmt.Sprintf("entries_%d", n), func(b *testing.B) {
			dir, err := os.MkdirTemp("", "bench-idx-rw-*")
			if err != nil {
				b.Fatal(err)
			}
			b.Cleanup(func() { _ = os.RemoveAll(dir) })

			gitDir := filepath.Join(dir, ".gitg")
			if err := os.MkdirAll(gitDir, 0755); err != nil {
				b.Fatal(err)
			}

			if err := Configure(WithPath(dir), WithGitDirectory(".gitg")); err != nil {
				b.Fatal(err)
			}

			idx := NewIndex()
			for i := 0; i < n; i++ {
				name := fmt.Sprintf("file%04d.txt", i)
				h := sha1.Sum([]byte(name))
				sha := Sha{set: true}
				copy(sha.hash[:], h[:])
				item := &indexItem{
					indexItemP: &indexItemP{
						CTimeS: uint32(time.Now().Unix()),
						MTimeS: uint32(time.Now().Unix()),
						Mode:   0100644,
						Size:   uint32(len(name)),
						Flags:  uint16(len(name)),
					},
					Name: []byte(name),
				}
				copy(item.indexItemP.Sha[:], h[:])
				idx.addItem(item)
			}
			if err := idx.Write(); err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := ReadIndex(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 5. Sha.AsHexString and Sha.Matches
// ---------------------------------------------------------------------------

func BenchmarkShaAsHexString(b *testing.B) {
	h := sha1.Sum([]byte("bench-hex-string"))
	s := Sha{set: true}
	copy(s.hash[:], h[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.AsHexString()
	}
}

func BenchmarkShaMatches(b *testing.B) {
	h1 := sha1.Sum([]byte("bench-match-a"))
	h2 := sha1.Sum([]byte("bench-match-a"))
	s1 := Sha{set: true}
	s2 := Sha{set: true}
	copy(s1.hash[:], h1[:])
	copy(s2.hash[:], h2[:])

	b.Run("equal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = s1.Matches(s2)
		}
	})

	h3 := sha1.Sum([]byte("bench-match-b"))
	s3 := Sha{set: true}
	copy(s3.hash[:], h3[:])

	b.Run("not_equal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = s1.Matches(s3)
		}
	})
}

// ---------------------------------------------------------------------------
// 6. NewFfileSet -- merging commit, index, and working tree file lists
// ---------------------------------------------------------------------------

func BenchmarkNewFfileSet(b *testing.B) {
	const n = 100
	dummySha := Sha{set: true}
	now := time.Now()

	mkFinfo := func(name string) *Finfo {
		return &Finfo{
			MTimeS: uint32(now.Unix()),
			MTimeN: 0,
			SSize:  42,
			NName:  name,
		}
	}

	commitFiles := make([]*FileStatus, n)
	indexFiles := make([]*FileStatus, n)
	wtFiles := make([]*FileStatus, n)

	for i := 0; i < n; i++ {
		name := fmt.Sprintf("path/to/file%04d.txt", i)
		commitFiles[i] = &FileStatus{
			path:   name,
			commit: &fileInfo{Sha: dummySha, Finfo: mkFinfo(name)},
		}
		indexFiles[i] = &FileStatus{
			path:  name,
			index: &fileInfo{Sha: dummySha, Finfo: mkFinfo(name)},
		}
		wtFiles[i] = &FileStatus{
			path: name,
			wd:   &fileInfo{Sha: dummySha, Finfo: mkFinfo(name)},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewFfileSet(commitFiles, indexFiles, wtFiles); err != nil {
			b.Fatal(err)
		}
	}
}

