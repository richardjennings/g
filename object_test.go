package g

import (
	"strings"
	"testing"
	"time"
)

// helper to create a Sha from a hex string, failing the test on error.
func mustSha(t *testing.T, hex string) Sha {
	t.Helper()
	sha, err := ShaFromHexString(hex)
	if err != nil {
		t.Fatalf("ShaFromHexString(%q): %v", hex, err)
	}
	return sha
}

const testWorkingDir = "/test/"

// helper to build a FileStatus with the given path and index sha.
func makeFileStatus(t *testing.T, relativePath string, shaHex string) *FileStatus {
	t.Helper()
	return &FileStatus{
		path: testWorkingDir + relativePath,
		index: &fileInfo{
			Sha: mustSha(t, shaHex),
		},
	}
}

func TestObjectTree(t *testing.T) {
	sha1Hex := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	sha2Hex := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	sha3Hex := "cccccccccccccccccccccccccccccccccccccccc"
	sha4Hex := "dddddddddddddddddddddddddddddddddddddddd"

	t.Run("empty file list", func(t *testing.T) {
		root := ObjectTree(nil, testWorkingDir)
		if root.Typ != ObjectTypeTree {
			t.Errorf("root type = %d, want ObjectTypeTree (%d)", root.Typ, ObjectTypeTree)
		}
		if len(root.Objects) != 0 {
			t.Errorf("root children = %d, want 0", len(root.Objects))
		}
	})

	t.Run("single top-level file", func(t *testing.T) {
		files := []*FileStatus{
			makeFileStatus(t, "readme.txt", sha1Hex),
		}
		root := ObjectTree(files, testWorkingDir)

		if root.Typ != ObjectTypeTree {
			t.Errorf("root type = %d, want ObjectTypeTree", root.Typ)
		}
		if len(root.Objects) != 1 {
			t.Fatalf("root children = %d, want 1", len(root.Objects))
		}
		child := root.Objects[0]
		if child.Typ != ObjectTypeBlob {
			t.Errorf("child type = %d, want ObjectTypeBlob (%d)", child.Typ, ObjectTypeBlob)
		}
		if !strings.HasSuffix(child.Path, "readme.txt") {
			t.Errorf("child path = %q, want suffix readme.txt", child.Path)
		}
		if child.Sha.AsHexString() != sha1Hex {
			t.Errorf("child sha = %s, want %s", child.Sha.AsHexString(), sha1Hex)
		}
	})

	t.Run("single file in subdirectory", func(t *testing.T) {
		files := []*FileStatus{
			makeFileStatus(t, "dir/file.txt", sha1Hex),
		}
		root := ObjectTree(files, testWorkingDir)

		if len(root.Objects) != 1 {
			t.Fatalf("root children = %d, want 1 (the dir tree)", len(root.Objects))
		}
		dirNode := root.Objects[0]
		if dirNode.Typ != ObjectTypeTree {
			t.Errorf("dir node type = %d, want ObjectTypeTree", dirNode.Typ)
		}
		if dirNode.Path != "dir" {
			t.Errorf("dir node path = %q, want %q", dirNode.Path, "dir")
		}
		if len(dirNode.Objects) != 1 {
			t.Fatalf("dir children = %d, want 1", len(dirNode.Objects))
		}
		fileNode := dirNode.Objects[0]
		if fileNode.Typ != ObjectTypeBlob {
			t.Errorf("file node type = %d, want ObjectTypeBlob", fileNode.Typ)
		}
		if fileNode.Sha.AsHexString() != sha1Hex {
			t.Errorf("file sha = %s, want %s", fileNode.Sha.AsHexString(), sha1Hex)
		}
	})

	t.Run("multiple files in same subdirectory exercises cache hit", func(t *testing.T) {
		files := []*FileStatus{
			makeFileStatus(t, "src/a.go", sha1Hex),
			makeFileStatus(t, "src/b.go", sha2Hex),
		}
		root := ObjectTree(files, testWorkingDir)

		if len(root.Objects) != 1 {
			t.Fatalf("root children = %d, want 1 (single src tree)", len(root.Objects))
		}
		srcNode := root.Objects[0]
		if srcNode.Typ != ObjectTypeTree {
			t.Errorf("src node type = %d, want ObjectTypeTree", srcNode.Typ)
		}
		if srcNode.Path != "src" {
			t.Errorf("src node path = %q, want %q", srcNode.Path, "src")
		}
		// both files should be children of the same src tree node
		if len(srcNode.Objects) != 2 {
			t.Fatalf("src children = %d, want 2", len(srcNode.Objects))
		}
		for i, want := range []string{sha1Hex, sha2Hex} {
			if srcNode.Objects[i].Typ != ObjectTypeBlob {
				t.Errorf("src child[%d] type = %d, want ObjectTypeBlob", i, srcNode.Objects[i].Typ)
			}
			if srcNode.Objects[i].Sha.AsHexString() != want {
				t.Errorf("src child[%d] sha = %s, want %s", i, srcNode.Objects[i].Sha.AsHexString(), want)
			}
		}
	})

	t.Run("nested directories", func(t *testing.T) {
		files := []*FileStatus{
			makeFileStatus(t, "a/b/file.txt", sha1Hex),
		}
		root := ObjectTree(files, testWorkingDir)

		if len(root.Objects) != 1 {
			t.Fatalf("root children = %d, want 1", len(root.Objects))
		}
		aNode := root.Objects[0]
		if aNode.Typ != ObjectTypeTree {
			t.Errorf("a node type = %d, want ObjectTypeTree", aNode.Typ)
		}
		if aNode.Path != "a" {
			t.Errorf("a node path = %q, want %q", aNode.Path, "a")
		}
		if len(aNode.Objects) != 1 {
			t.Fatalf("a children = %d, want 1", len(aNode.Objects))
		}
		bNode := aNode.Objects[0]
		if bNode.Typ != ObjectTypeTree {
			t.Errorf("b node type = %d, want ObjectTypeTree", bNode.Typ)
		}
		if bNode.Path != "b" {
			t.Errorf("b node path = %q, want %q", bNode.Path, "b")
		}
		if len(bNode.Objects) != 1 {
			t.Fatalf("b children = %d, want 1", len(bNode.Objects))
		}
		fileNode := bNode.Objects[0]
		if fileNode.Typ != ObjectTypeBlob {
			t.Errorf("file node type = %d, want ObjectTypeBlob", fileNode.Typ)
		}
		if fileNode.Sha.AsHexString() != sha1Hex {
			t.Errorf("file sha = %s, want %s", fileNode.Sha.AsHexString(), sha1Hex)
		}
	})

	t.Run("mix of top-level and nested files", func(t *testing.T) {
		files := []*FileStatus{
			makeFileStatus(t, "root.txt", sha1Hex),
			makeFileStatus(t, "dir/nested.txt", sha2Hex),
			makeFileStatus(t, "dir/other.txt", sha3Hex),
			makeFileStatus(t, "deep/sub/leaf.txt", sha4Hex),
		}
		root := ObjectTree(files, testWorkingDir)

		// root should have 3 children: root.txt blob, dir tree, deep tree
		if len(root.Objects) != 3 {
			t.Fatalf("root children = %d, want 3", len(root.Objects))
		}

		// first child: top-level blob
		if root.Objects[0].Typ != ObjectTypeBlob {
			t.Errorf("root.Objects[0] type = %d, want ObjectTypeBlob", root.Objects[0].Typ)
		}
		if root.Objects[0].Sha.AsHexString() != sha1Hex {
			t.Errorf("root.Objects[0] sha = %s, want %s", root.Objects[0].Sha.AsHexString(), sha1Hex)
		}

		// second child: dir tree with 2 blobs
		dirNode := root.Objects[1]
		if dirNode.Typ != ObjectTypeTree {
			t.Errorf("dir node type = %d, want ObjectTypeTree", dirNode.Typ)
		}
		if dirNode.Path != "dir" {
			t.Errorf("dir node path = %q, want %q", dirNode.Path, "dir")
		}
		if len(dirNode.Objects) != 2 {
			t.Fatalf("dir children = %d, want 2", len(dirNode.Objects))
		}

		// third child: deep tree -> sub tree -> leaf blob
		deepNode := root.Objects[2]
		if deepNode.Typ != ObjectTypeTree {
			t.Errorf("deep node type = %d, want ObjectTypeTree", deepNode.Typ)
		}
		if deepNode.Path != "deep" {
			t.Errorf("deep node path = %q, want %q", deepNode.Path, "deep")
		}
		if len(deepNode.Objects) != 1 {
			t.Fatalf("deep children = %d, want 1", len(deepNode.Objects))
		}
		subNode := deepNode.Objects[0]
		if subNode.Typ != ObjectTypeTree {
			t.Errorf("sub node type = %d, want ObjectTypeTree", subNode.Typ)
		}
		if subNode.Path != "sub" {
			t.Errorf("sub node path = %q, want %q", subNode.Path, "sub")
		}
		if len(subNode.Objects) != 1 {
			t.Fatalf("sub children = %d, want 1", len(subNode.Objects))
		}
		leafNode := subNode.Objects[0]
		if leafNode.Typ != ObjectTypeBlob {
			t.Errorf("leaf node type = %d, want ObjectTypeBlob", leafNode.Typ)
		}
		if leafNode.Sha.AsHexString() != sha4Hex {
			t.Errorf("leaf sha = %s, want %s", leafNode.Sha.AsHexString(), sha4Hex)
		}
	})
}

func TestCommitString(t *testing.T) {
	commitSha := mustSha(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	treeSha := mustSha(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	parent1 := mustSha(t, "cccccccccccccccccccccccccccccccccccccccc")
	parent2 := mustSha(t, "dddddddddddddddddddddddddddddddddddddddd")

	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	t.Run("no parents", func(t *testing.T) {
		c := Commit{
			Sha:            commitSha,
			Tree:           treeSha,
			Author:         "Alice",
			AuthorEmail:    "alice@example.com",
			AuthoredTime:   fixedTime,
			Committer:      "Bob",
			CommitterEmail: "bob@example.com",
			CommittedTime:  fixedTime,
			Message:        []byte("initial commit"),
		}
		out := c.String()

		if !strings.Contains(out, "commit: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
			t.Errorf("missing commit sha in output:\n%s", out)
		}
		if !strings.Contains(out, "tree: bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb") {
			t.Errorf("missing tree sha in output:\n%s", out)
		}
		if strings.Contains(out, "parent:") {
			t.Errorf("unexpected parent line in output:\n%s", out)
		}
		if !strings.Contains(out, "Alice <alice@example.com>") {
			t.Errorf("missing author in output:\n%s", out)
		}
		if !strings.Contains(out, "Bob <bob@example.com>") {
			t.Errorf("missing committer in output:\n%s", out)
		}
		if !strings.Contains(out, "message: \ninitial commit") {
			t.Errorf("missing message in output:\n%s", out)
		}
	})

	t.Run("one parent", func(t *testing.T) {
		c := Commit{
			Sha:            commitSha,
			Tree:           treeSha,
			Parents:        []Sha{parent1},
			Author:         "Alice",
			AuthorEmail:    "alice@example.com",
			AuthoredTime:   fixedTime,
			Committer:      "Bob",
			CommitterEmail: "bob@example.com",
			CommittedTime:  fixedTime,
			Message:        []byte("second commit"),
		}
		out := c.String()

		if !strings.Contains(out, "parent: cccccccccccccccccccccccccccccccccccccccc") {
			t.Errorf("missing parent sha in output:\n%s", out)
		}
		// should have exactly one parent line
		if strings.Count(out, "parent:") != 1 {
			t.Errorf("expected 1 parent line, got %d in output:\n%s", strings.Count(out, "parent:"), out)
		}
	})

	t.Run("two parents (merge commit)", func(t *testing.T) {
		c := Commit{
			Sha:            commitSha,
			Tree:           treeSha,
			Parents:        []Sha{parent1, parent2},
			Author:         "Alice",
			AuthorEmail:    "alice@example.com",
			AuthoredTime:   fixedTime,
			Committer:      "Bob",
			CommitterEmail: "bob@example.com",
			CommittedTime:  fixedTime,
			Message:        []byte("merge branch"),
		}
		out := c.String()

		if !strings.Contains(out, "parent: cccccccccccccccccccccccccccccccccccccccc") {
			t.Errorf("missing first parent sha in output:\n%s", out)
		}
		if !strings.Contains(out, "parent: dddddddddddddddddddddddddddddddddddddddd") {
			t.Errorf("missing second parent sha in output:\n%s", out)
		}
		if strings.Count(out, "parent:") != 2 {
			t.Errorf("expected 2 parent lines, got %d in output:\n%s", strings.Count(out, "parent:"), out)
		}
		if !strings.Contains(out, "message: \nmerge branch") {
			t.Errorf("missing message in output:\n%s", out)
		}
	})
}
