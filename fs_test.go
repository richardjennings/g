package g

import (
	"fmt"
	"os"
	"testing"
)

func TestIndexStatus_StatusString(t *testing.T) {
	tests := []struct {
		status IndexStatus
		want   string
	}{
		{NotUpdated, " "},
		{UpdatedInIndex, "M"},
		{TypeChangedInIndex, "T"},
		{AddedInIndex, "A"},
		{DeletedInIndex, "D"},
		{RenamedInIndex, "R"},
		{CopiedInIndex, "C"},
		{UntrackedInIndex, "?"},
		{IndexStatus(255), ""},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("IndexStatus(%d)", tt.status), func(t *testing.T) {
			got := tt.status.StatusString()
			if got != tt.want {
				t.Errorf("IndexStatus(%d).StatusString() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestWDStatus_StatusString(t *testing.T) {
	tests := []struct {
		status WDStatus
		want   string
	}{
		{IndexAndWorkingTreeMatch, " "},
		{WorktreeChangedSinceIndex, "M"},
		{TypeChangedInWorktreeSinceIndex, "T"},
		{DeletedInWorktree, "D"},
		{RenamedInWorktree, "R"},
		{CopiedInWorktree, "C"},
		{Untracked, "?"},
		{WDStatus(255), ""},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("WDStatus(%d)", tt.status), func(t *testing.T) {
			got := tt.status.StatusString()
			if got != tt.want {
				t.Errorf("WDStatus(%d).StatusString() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestIndexStatus_String(t *testing.T) {
	tests := []struct {
		status IndexStatus
		want   string
	}{
		{NotUpdated, "NotUpdated"},
		{UpdatedInIndex, "UpdatedInIndex"},
		{TypeChangedInIndex, "TypeChangedInIndex"},
		{AddedInIndex, "AddedInIndex"},
		{DeletedInIndex, "DeletedInIndex"},
		{RenamedInIndex, "RenamedInIndex"},
		{CopiedInIndex, "CopiedInIndex"},
		{UntrackedInIndex, "UntrackedInIndex"},
		{IndexStatus(255), "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("IndexStatus(%d)", tt.status), func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("IndexStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestWDStatus_String(t *testing.T) {
	tests := []struct {
		status WDStatus
		want   string
	}{
		{IndexAndWorkingTreeMatch, "IndexAndWorkingTreeMatch"},
		{WorktreeChangedSinceIndex, "WorktreeChangedSinceIndex"},
		{TypeChangedInWorktreeSinceIndex, "TypeChangedInWorktreeSinceIndex"},
		{DeletedInWorktree, "DeletedInWorktree"},
		{RenamedInWorktree, "RenamedInWorktree"},
		{CopiedInWorktree, "CopiedInWorktree"},
		{Untracked, "Untracked"},
		{WDStatus(255), "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("WDStatus(%d)", tt.status), func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("WDStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestNewFfileSet_UpdateStatus(t *testing.T) {
	// Helper to create a Sha with a specific first byte (rest zero).
	makeSha := func(b byte) Sha {
		s, _ := NewSha([]byte{b, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		return s
	}

	// Helper to create a minimal Finfo acting as os.FileInfo with a given mtime.
	makeFinfo := func(sec uint32) *Finfo {
		return &Finfo{MTimeS: sec}
	}

	sha1 := makeSha(1)
	sha2 := makeSha(2)

	t.Run("commit only - deleted from index and worktree", func(t *testing.T) {
		// File exists only in commit (not in index, not in wd).
		// Expected: wdStatus = IndexAndWorkingTreeMatch, idxStatus = DeletedInIndex.
		commitFiles := []*FileStatus{{path: "a.txt", commit: &fileInfo{Sha: sha1, Finfo: makeFinfo(100)}}}
		fs, err := NewFfileSet(commitFiles, nil, nil)
		if err != nil {
			t.Fatalf("NewFfileSet error: %v", err)
		}
		f := fs.Files()[0]
		if f.WorkingDirectoryStatus() != IndexAndWorkingTreeMatch {
			t.Errorf("wdStatus = %v, want IndexAndWorkingTreeMatch", f.WorkingDirectoryStatus())
		}
		if f.IndexStatus() != DeletedInIndex {
			t.Errorf("idxStatus = %v, want DeletedInIndex", f.IndexStatus())
		}
	})

	t.Run("commit and index with different SHAs", func(t *testing.T) {
		// File in commit and index with different SHAs, and index in wd with same mtime.
		// Expected: wdStatus = IndexAndWorkingTreeMatch, idxStatus = UpdatedInIndex.
		commitFiles := []*FileStatus{{path: "b.txt", commit: &fileInfo{Sha: sha1, Finfo: makeFinfo(100)}}}
		indexFiles := []*FileStatus{{path: "b.txt", index: &fileInfo{Sha: sha2, Finfo: makeFinfo(200)}}}
		wdFiles := []*FileStatus{{path: "b.txt", wd: &fileInfo{Finfo: makeFinfo(200)}}}
		fs, err := NewFfileSet(commitFiles, indexFiles, wdFiles)
		if err != nil {
			t.Fatalf("NewFfileSet error: %v", err)
		}
		f := fs.Files()[0]
		if f.WorkingDirectoryStatus() != IndexAndWorkingTreeMatch {
			t.Errorf("wdStatus = %v, want IndexAndWorkingTreeMatch", f.WorkingDirectoryStatus())
		}
		if f.IndexStatus() != UpdatedInIndex {
			t.Errorf("idxStatus = %v, want UpdatedInIndex", f.IndexStatus())
		}
	})

	t.Run("commit and index with same SHA", func(t *testing.T) {
		// File in commit and index with same SHA, and in wd with different mtime.
		// Expected: wdStatus = WorktreeChangedSinceIndex, idxStatus = NotUpdated.
		commitFiles := []*FileStatus{{path: "c.txt", commit: &fileInfo{Sha: sha1, Finfo: makeFinfo(100)}}}
		indexFiles := []*FileStatus{{path: "c.txt", index: &fileInfo{Sha: sha1, Finfo: makeFinfo(100)}}}
		wdFiles := []*FileStatus{{path: "c.txt", wd: &fileInfo{Finfo: makeFinfo(999)}}}
		fs, err := NewFfileSet(commitFiles, indexFiles, wdFiles)
		if err != nil {
			t.Fatalf("NewFfileSet error: %v", err)
		}
		f := fs.Files()[0]
		if f.WorkingDirectoryStatus() != WorktreeChangedSinceIndex {
			t.Errorf("wdStatus = %v, want WorktreeChangedSinceIndex", f.WorkingDirectoryStatus())
		}
		if f.IndexStatus() != NotUpdated {
			t.Errorf("idxStatus = %v, want NotUpdated", f.IndexStatus())
		}
	})

	t.Run("index only - added in index, deleted in worktree", func(t *testing.T) {
		// File only in index (not in commit, not in wd).
		// Expected: wdStatus = DeletedInWorktree, idxStatus = AddedInIndex.
		indexFiles := []*FileStatus{{path: "d.txt", index: &fileInfo{Sha: sha1, Finfo: makeFinfo(100)}}}
		fs, err := NewFfileSet(nil, indexFiles, nil)
		if err != nil {
			t.Fatalf("NewFfileSet error: %v", err)
		}
		f := fs.Files()[0]
		if f.WorkingDirectoryStatus() != DeletedInWorktree {
			t.Errorf("wdStatus = %v, want DeletedInWorktree", f.WorkingDirectoryStatus())
		}
		if f.IndexStatus() != AddedInIndex {
			t.Errorf("idxStatus = %v, want AddedInIndex", f.IndexStatus())
		}
	})

	t.Run("wd only - untracked", func(t *testing.T) {
		// File only in working directory (not in commit or index).
		// Expected: wdStatus = Untracked, idxStatus = UntrackedInIndex.
		wdFiles := []*FileStatus{{path: "e.txt", wd: &fileInfo{Finfo: makeFinfo(100)}}}
		fs, err := NewFfileSet(nil, nil, wdFiles)
		if err != nil {
			t.Fatalf("NewFfileSet error: %v", err)
		}
		f := fs.Files()[0]
		if f.WorkingDirectoryStatus() != Untracked {
			t.Errorf("wdStatus = %v, want Untracked", f.WorkingDirectoryStatus())
		}
		if f.IndexStatus() != UntrackedInIndex {
			t.Errorf("idxStatus = %v, want UntrackedInIndex", f.IndexStatus())
		}
	})

	t.Run("Contains and Files", func(t *testing.T) {
		commitFiles := []*FileStatus{{path: "f.txt", commit: &fileInfo{Sha: sha1, Finfo: makeFinfo(100)}}}
		indexFiles := []*FileStatus{{path: "f.txt", index: &fileInfo{Sha: sha1, Finfo: makeFinfo(100)}}}
		wdFiles := []*FileStatus{{path: "f.txt", wd: &fileInfo{Finfo: makeFinfo(100)}}}
		fs, err := NewFfileSet(commitFiles, indexFiles, wdFiles)
		if err != nil {
			t.Fatalf("NewFfileSet error: %v", err)
		}
		if _, ok := fs.Contains("f.txt"); !ok {
			t.Error("Contains(f.txt) = false, want true")
		}
		if _, ok := fs.Contains("nonexistent.txt"); ok {
			t.Error("Contains(nonexistent.txt) = true, want false")
		}
		if len(fs.Files()) != 1 {
			t.Errorf("Files() length = %d, want 1", len(fs.Files()))
		}
	})

	t.Run("Path accessor", func(t *testing.T) {
		wdFiles := []*FileStatus{{path: "hello.txt", wd: &fileInfo{Finfo: makeFinfo(100)}}}
		fs, err := NewFfileSet(nil, nil, wdFiles)
		if err != nil {
			t.Fatalf("NewFfileSet error: %v", err)
		}
		if fs.Files()[0].Path() != "hello.txt" {
			t.Errorf("Path() = %q, want %q", fs.Files()[0].Path(), "hello.txt")
		}
	})
}

func TestFinfo_Methods(t *testing.T) {
	fi := &Finfo{
		NName: "hello.txt",
		SSize: 42,
		MMode: 0100644,
	}

	t.Run("Name", func(t *testing.T) {
		if got := fi.Name(); got != "hello.txt" {
			t.Errorf("Name() = %q, want %q", got, "hello.txt")
		}
	})

	t.Run("Size", func(t *testing.T) {
		if got := fi.Size(); got != 42 {
			t.Errorf("Size() = %d, want %d", got, 42)
		}
	})

	t.Run("Mode", func(t *testing.T) {
		if got := fi.Mode(); got != os.FileMode(0) {
			t.Errorf("Mode() = %v, want %v", got, os.FileMode(0))
		}
	})

	t.Run("IsDir", func(t *testing.T) {
		if got := fi.IsDir(); got != false {
			t.Errorf("IsDir() = %v, want false", got)
		}
	})

	t.Run("Sys", func(t *testing.T) {
		if got := fi.Sys(); got != nil {
			t.Errorf("Sys() = %v, want nil", got)
		}
	})
}
