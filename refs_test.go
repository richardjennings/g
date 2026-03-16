package g

import (
	"os"
	"path/filepath"
	"testing"
)

const testSHA = "aabbccddee0011223344aabbccddee0011223344"

// setupGitgDir creates a minimal .gitg directory structure in a temp directory
// and configures the global config to point at it. It returns the temp dir path.
func setupGitgDir(t *testing.T, branches map[string]string, headBranch string) string {
	t.Helper()
	tmp := t.TempDir()

	gitgDir := filepath.Join(tmp, ".gitg")
	refsHeads := filepath.Join(gitgDir, "refs", "heads")
	infoDir := filepath.Join(gitgDir, "info")

	if err := os.MkdirAll(refsHeads, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write branch files
	for name, sha := range branches {
		branchFile := filepath.Join(refsHeads, name)
		if err := os.WriteFile(branchFile, []byte(sha+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Write HEAD
	headContent := "ref: refs/heads/" + headBranch + "\n"
	if err := os.WriteFile(filepath.Join(gitgDir, "HEAD"), []byte(headContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write empty info/refs (packed-refs) so packedrefs() doesn't fail
	if err := os.WriteFile(filepath.Join(infoDir, "refs"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Configure(WithPath(tmp), WithGitDirectory(".gitg")); err != nil {
		t.Fatal(err)
	}

	return tmp
}

func TestListBranches(t *testing.T) {
	setupGitgDir(t, map[string]string{
		"main":    testSHA,
		"develop": testSHA,
		"feature": testSHA,
	}, "main")

	branches, err := ListBranches()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"develop", "feature", "main"}
	if len(branches) != len(expected) {
		t.Fatalf("len(branches) = %d, want %d", len(branches), len(expected))
	}
	for i, name := range expected {
		if branches[i] != name {
			t.Errorf("branches[%d] = %q, want %q", i, branches[i], name)
		}
	}
}

func TestListBranches_Single(t *testing.T) {
	setupGitgDir(t, map[string]string{
		"main": testSHA,
	}, "main")

	branches, err := ListBranches()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("len(branches) = %d, want 1", len(branches))
	}
	if branches[0] != "main" {
		t.Errorf("branches[0] = %q, want %q", branches[0], "main")
	}
}

func TestDeleteBranch(t *testing.T) {
	setupGitgDir(t, map[string]string{
		"main":    testSHA,
		"to-delete": testSHA,
	}, "main")

	// Verify the branch file exists before deletion
	branchFile := filepath.Join(RefsHeadsDirectory(), "to-delete")
	if _, err := os.Stat(branchFile); err != nil {
		t.Fatalf("branch file should exist before delete: %v", err)
	}

	err := DeleteBranch("to-delete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the file is gone
	if _, err := os.Stat(branchFile); !os.IsNotExist(err) {
		t.Error("expected branch file to be removed after DeleteBranch")
	}
}

func TestDeleteBranch_NonExistent(t *testing.T) {
	setupGitgDir(t, map[string]string{
		"main": testSHA,
	}, "main")

	err := DeleteBranch("no-such-branch")
	if err == nil {
		t.Fatal("expected error when deleting non-existent branch")
	}
}

func TestCreateBranch(t *testing.T) {
	setupGitgDir(t, map[string]string{
		"main": testSHA,
	}, "main")

	err := CreateBranch("new-feature")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the new branch file was created
	branchFile := filepath.Join(RefsHeadsDirectory(), "new-feature")
	data, err := os.ReadFile(branchFile)
	if err != nil {
		t.Fatalf("could not read new branch file: %v", err)
	}

	// The new branch should point to the same SHA as main
	sha, err := NewSha([]byte(testSHA))
	if err != nil {
		t.Fatal(err)
	}
	// The file content should be the hex SHA + newline
	expected := sha.AsHexString() + "\n"
	if string(data) != expected {
		t.Errorf("branch file content = %q, want %q", string(data), expected)
	}
}

func TestCreateBranch_VerifyInList(t *testing.T) {
	setupGitgDir(t, map[string]string{
		"main": testSHA,
	}, "main")

	err := CreateBranch("alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	branches, err := ListBranches()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should now have both "alpha" and "main", sorted
	expected := []string{"alpha", "main"}
	if len(branches) != len(expected) {
		t.Fatalf("len(branches) = %d, want %d", len(branches), len(expected))
	}
	for i, name := range expected {
		if branches[i] != name {
			t.Errorf("branches[%d] = %q, want %q", i, branches[i], name)
		}
	}
}
