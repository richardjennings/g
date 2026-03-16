package g

import (
	"testing"
)

// helper to build an Index with the given file paths, each with a distinct valid SHA.
func testIndex(paths ...string) *Index {
	idx := NewIndex()
	for i, p := range paths {
		var sha [20]byte
		sha[0] = byte(i + 1) // non-zero so NewSha succeeds on the 20-byte slice
		item := &indexItem{
			indexItemP: &indexItemP{
				Sha:  sha,
				Mode: 0100644,
			},
			Name: []byte(p),
		}
		idx.items = append(idx.items, item)
		idx.header.NumEntries++
	}
	return idx
}

func TestIndex_File_Found(t *testing.T) {
	idx := testIndex("hello.txt", "world.txt", "foo/bar.go")

	fs, err := idx.File("world.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs == nil {
		t.Fatal("expected non-nil FileStatus for existing path")
	}
	if fs.Path() != "world.txt" {
		t.Errorf("path = %q, want %q", fs.Path(), "world.txt")
	}
}

func TestIndex_File_NotFound(t *testing.T) {
	idx := testIndex("hello.txt")

	fs, err := idx.File("missing.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs != nil {
		t.Errorf("expected nil FileStatus for non-existent path, got %+v", fs)
	}
}

func TestIndex_File_EmptyIndex(t *testing.T) {
	idx := NewIndex()

	fs, err := idx.File("anything.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs != nil {
		t.Errorf("expected nil for empty index, got %+v", fs)
	}
}

func TestIndex_Rm_Existing(t *testing.T) {
	idx := testIndex("a.txt", "b.txt", "c.txt")

	if idx.header.NumEntries != 3 {
		t.Fatalf("NumEntries = %d, want 3", idx.header.NumEntries)
	}

	err := idx.Rm("b.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx.header.NumEntries != 2 {
		t.Errorf("NumEntries after Rm = %d, want 2", idx.header.NumEntries)
	}
	if len(idx.items) != 2 {
		t.Errorf("len(items) after Rm = %d, want 2", len(idx.items))
	}

	// Verify the correct item was removed
	for _, item := range idx.items {
		if string(item.Name) == "b.txt" {
			t.Error("b.txt should have been removed from items")
		}
	}
}

func TestIndex_Rm_NonExistent(t *testing.T) {
	idx := testIndex("a.txt")

	err := idx.Rm("nope.txt")
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
	// NumEntries should be unchanged
	if idx.header.NumEntries != 1 {
		t.Errorf("NumEntries = %d, want 1", idx.header.NumEntries)
	}
}

func TestIndex_Rm_FirstItem(t *testing.T) {
	idx := testIndex("first.txt", "second.txt")

	err := idx.Rm("first.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx.header.NumEntries != 1 {
		t.Errorf("NumEntries = %d, want 1", idx.header.NumEntries)
	}
	if string(idx.items[0].Name) != "second.txt" {
		t.Errorf("remaining item = %q, want %q", string(idx.items[0].Name), "second.txt")
	}
}

func TestIndex_Rm_LastItem(t *testing.T) {
	idx := testIndex("first.txt", "second.txt")

	err := idx.Rm("second.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx.header.NumEntries != 1 {
		t.Errorf("NumEntries = %d, want 1", idx.header.NumEntries)
	}
	if string(idx.items[0].Name) != "first.txt" {
		t.Errorf("remaining item = %q, want %q", string(idx.items[0].Name), "first.txt")
	}
}

func TestIndex_Rm_OnlyItem(t *testing.T) {
	idx := testIndex("only.txt")

	err := idx.Rm("only.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx.header.NumEntries != 0 {
		t.Errorf("NumEntries = %d, want 0", idx.header.NumEntries)
	}
	if len(idx.items) != 0 {
		t.Errorf("len(items) = %d, want 0", len(idx.items))
	}
}

func TestIndex_Files(t *testing.T) {
	idx := testIndex("a.txt", "b.txt")

	files, err := idx.Files()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("len(files) = %d, want 2", len(files))
	}
	if files[0].Path() != "a.txt" {
		t.Errorf("files[0].Path() = %q, want %q", files[0].Path(), "a.txt")
	}
	if files[1].Path() != "b.txt" {
		t.Errorf("files[1].Path() = %q, want %q", files[1].Path(), "b.txt")
	}
}

func TestIndex_Files_Empty(t *testing.T) {
	idx := NewIndex()

	files, err := idx.Files()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("len(files) = %d, want 0", len(files))
	}
}
