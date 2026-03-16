package g

import (
	"encoding/binary"
	"os"
	"testing"
)

func TestPackfile_lookupInPackfiles(t *testing.T) {
	if err := Configure(
		WithPath("./test_assets/repo/test-pack-file"),
		WithGitDirectory(".gitg"),
	); err != nil {
		t.Fatal(err)
	}
	sha, err := ShaFromHexString("d78ccc12bfbd1e6e0a53a9dd503cdec24f1866d6")
	if err != nil {
		t.Fatal(err)
	}
	obj, err := lookupInPackfiles(sha)
	if err != nil {
		t.Fatal(err)
	}
	if obj == nil {
		t.Fatal("expected non-nil object")
	}
	if obj.Typ != ObjectTypeCommit {
		t.Errorf("typ = %d, want %d", obj.Typ, ObjectTypeCommit)
	}
	files, err := CurrentStatus()
	if err != nil {
		t.Fatal(err)
	}
	if files.Files()[0].idxStatus != NotUpdated {
		t.Errorf("idxStatus = %d, want %d", files.Files()[0].idxStatus, NotUpdated)
	}
}

// writeSortedHashes writes sorted 20-byte hashes to a temp file and returns
// the open file seeked to the start, ready for findObjectName.
func writeSortedHashes(t *testing.T, hashes [][20]byte) *os.File {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "idx-test-*")
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range hashes {
		if err := binary.Write(f, binary.BigEndian, h); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestFindObjectName(t *testing.T) {
	// 5 sorted hashes to exercise binary search
	hashes := [][20]byte{
		{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		{0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
		{0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03},
		{0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04},
		{0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05},
	}

	tests := []struct {
		name      string
		target    [20]byte
		wantIdx   uint32
		wantFound bool
	}{
		{
			name:      "find first entry",
			target:    hashes[0],
			wantIdx:   0,
			wantFound: true,
		},
		{
			name:      "find middle entry",
			target:    hashes[2],
			wantIdx:   2,
			wantFound: true,
		},
		{
			name:      "find last entry",
			target:    hashes[4],
			wantIdx:   4,
			wantFound: true,
		},
		{
			name:      "not found - before all",
			target:    [20]byte{0x05},
			wantIdx:   0,
			wantFound: false,
		},
		{
			name:      "not found - between entries",
			target:    [20]byte{0x25},
			wantIdx:   0,
			wantFound: false,
		},
		{
			name:      "not found - after all",
			target:    [20]byte{0xFF},
			wantIdx:   0,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := writeSortedHashes(t, hashes)
			defer f.Close()

			sha := Sha{set: true, hash: tt.target}
			idx, found, err := findObjectName(uint32(len(hashes)), f, sha)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
			if found && idx != tt.wantIdx {
				t.Errorf("idx = %d, want %d", idx, tt.wantIdx)
			}
		})
	}

	t.Run("empty bucket", func(t *testing.T) {
		f := writeSortedHashes(t, nil)
		defer f.Close()

		sha := Sha{set: true, hash: [20]byte{0x10}}
		_, found, err := findObjectName(0, f, sha)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found {
			t.Error("expected not found for empty bucket")
		}
	})

	t.Run("single entry found", func(t *testing.T) {
		f := writeSortedHashes(t, hashes[2:3])
		defer f.Close()

		sha := Sha{set: true, hash: hashes[2]}
		idx, found, err := findObjectName(1, f, sha)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !found {
			t.Fatal("expected found")
		}
		if idx != 0 {
			t.Errorf("idx = %d, want 0", idx)
		}
	})

	t.Run("single entry not found", func(t *testing.T) {
		f := writeSortedHashes(t, hashes[2:3])
		defer f.Close()

		sha := Sha{set: true, hash: [20]byte{0x35}}
		_, found, err := findObjectName(1, f, sha)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found {
			t.Error("expected not found")
		}
	})
}
