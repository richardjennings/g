package g

import (
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
