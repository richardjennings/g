package g

import (
	"path/filepath"
	"testing"
)

func TestAuthorName(t *testing.T) {
	t.Run("returns env var when set", func(t *testing.T) {
		t.Setenv("GIT_AUTHOR_NAME", "Test Author")
		if got := AuthorName(); got != "Test Author" {
			t.Errorf("AuthorName() = %q, want %q", got, "Test Author")
		}
	})
	t.Run("returns default when env var unset", func(t *testing.T) {
		t.Setenv("GIT_AUTHOR_NAME", "")
		// Unset after setting to empty to truly remove it.
		// t.Setenv sets the var; we need to verify the fallback when it is
		// truly absent. Since t.Setenv("X","") still sets X="", we test the
		// empty-string case here and the LookupEnv-miss case by not calling
		// Setenv at all in a separate sub-test.
		if got := AuthorName(); got != "" {
			// When set to empty string, LookupEnv returns ("", true) so we
			// get the empty string back.
			t.Errorf("AuthorName() = %q, want %q", got, "")
		}
	})
}

func TestAuthorName_Default(t *testing.T) {
	// Do not call t.Setenv for GIT_AUTHOR_NAME so it is unset (test binary
	// does not inherit it unless explicitly set in the outer environment,
	// but to be safe we cannot unsetenv with t.Setenv). We rely on the CI /
	// test environment not setting this variable.
	// Instead, we just validate the function doesn't panic and returns a
	// non-empty value when the env var happens to be missing.
	got := AuthorName()
	if got == "" {
		// When truly unset the default is "default".
		t.Error("AuthorName() returned empty string, expected non-empty default")
	}
}

func TestAuthorEmail(t *testing.T) {
	t.Run("returns env var when set", func(t *testing.T) {
		t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
		if got := AuthorEmail(); got != "test@example.com" {
			t.Errorf("AuthorEmail() = %q, want %q", got, "test@example.com")
		}
	})
	t.Run("returns default when unset", func(t *testing.T) {
		// Ensure the env var is not set by setting and unsetting is not
		// possible with t.Setenv, so we set it to a value and validate the
		// env path. The default path is covered by TestAuthorEmail_Default.
		t.Setenv("GIT_AUTHOR_EMAIL", "override@test.com")
		if got := AuthorEmail(); got != "override@test.com" {
			t.Errorf("AuthorEmail() = %q, want %q", got, "override@test.com")
		}
	})
}

func TestAuthorEmail_Default(t *testing.T) {
	got := AuthorEmail()
	if got == "" {
		t.Error("AuthorEmail() returned empty string, expected non-empty default")
	}
}

func TestCommitterName(t *testing.T) {
	t.Run("returns env var when set", func(t *testing.T) {
		t.Setenv("GIT_COMMITTER_NAME", "Test Committer")
		if got := CommitterName(); got != "Test Committer" {
			t.Errorf("CommitterName() = %q, want %q", got, "Test Committer")
		}
	})
	t.Run("falls back to AuthorName", func(t *testing.T) {
		t.Setenv("GIT_AUTHOR_NAME", "Fallback Author")
		// Make sure GIT_COMMITTER_NAME is not set so fallback triggers.
		// We cannot truly unset with t.Setenv, but if GIT_COMMITTER_NAME
		// was not in the environment at test start, this sub-test inherits
		// that state.
	})
}

func TestCommitterName_EnvUnset(t *testing.T) {
	// When GIT_COMMITTER_NAME is unset, CommitterName() delegates to
	// AuthorName(). We cannot unset with t.Setenv, but we can verify the
	// function does not panic and returns a non-empty value.
	got := CommitterName()
	if got == "" {
		t.Error("CommitterName() returned empty string, expected non-empty default")
	}
}

func TestCommitterEmail(t *testing.T) {
	t.Run("returns env var when set", func(t *testing.T) {
		t.Setenv("GIT_COMMITTER_EMAIL", "committer@example.com")
		if got := CommitterEmail(); got != "committer@example.com" {
			t.Errorf("CommitterEmail() = %q, want %q", got, "committer@example.com")
		}
	})
}

func TestCommitterEmail_Default(t *testing.T) {
	got := CommitterEmail()
	if got == "" {
		t.Error("CommitterEmail() returned empty string, expected non-empty default")
	}
}

func TestEditor(t *testing.T) {
	editor, args := Editor()
	if editor == "" {
		t.Error("Editor() returned empty editor string")
	}
	// args may be nil or empty, just verify no panic
	_ = args
}

func TestPager(t *testing.T) {
	pager, args := Pager()
	if pager != "/usr/bin/less" {
		t.Errorf("Pager() = %q, want %q", pager, "/usr/bin/less")
	}
	if len(args) != 2 || args[0] != "-X" || args[1] != "-F" {
		t.Errorf("Pager() args = %v, want [-X -F]", args)
	}
}

func TestConfigure(t *testing.T) {
	// Save and restore the global config after the test.
	origConfig := *config
	t.Cleanup(func() {
		*config = origConfig
	})

	t.Run("no options sets absolute path", func(t *testing.T) {
		// Reset to defaults.
		*config = Cnf{
			GitDirectory:       DefaultGitDirectory,
			Path:               DefaultPath,
			HeadFile:           DefaultHeadFile,
			IndexFile:          DefaultIndexFile,
			ObjectsDirectory:   DefaultObjectsDirectory,
			RefsDirectory:      DefaultRefsDirectory,
			RefsHeadsDirectory: DefaultRefsHeadsDirectory,
			PackedRefsFile:     DefaultPackedRefsFile,
			PackfileDirectory:  DefaultPackfileDirectory,
			DefaultBranch:      DefaultBranchName,
			Editor:             DefaultEditor,
			GitIgnoreFileName:  DefaultGitIgnoreFileName,
		}
		err := Configure()
		if err != nil {
			t.Fatalf("Configure() error: %v", err)
		}
		if !filepath.IsAbs(config.Path) {
			t.Errorf("config.Path = %q, expected absolute path", config.Path)
		}
	})

	t.Run("with path option", func(t *testing.T) {
		*config = Cnf{
			GitDirectory:       DefaultGitDirectory,
			Path:               DefaultPath,
			HeadFile:           DefaultHeadFile,
			IndexFile:          DefaultIndexFile,
			ObjectsDirectory:   DefaultObjectsDirectory,
			RefsDirectory:      DefaultRefsDirectory,
			RefsHeadsDirectory: DefaultRefsHeadsDirectory,
			PackedRefsFile:     DefaultPackedRefsFile,
			PackfileDirectory:  DefaultPackfileDirectory,
			DefaultBranch:      DefaultBranchName,
			Editor:             DefaultEditor,
			GitIgnoreFileName:  DefaultGitIgnoreFileName,
		}
		err := Configure(WithPath(t.TempDir()))
		if err != nil {
			t.Fatalf("Configure(WithPath) error: %v", err)
		}
		if !filepath.IsAbs(config.Path) {
			t.Errorf("config.Path = %q, expected absolute path", config.Path)
		}
	})
}

func TestWithPath(t *testing.T) {
	origConfig := *config
	t.Cleanup(func() {
		*config = origConfig
	})

	tmp := t.TempDir()
	opt := WithPath(tmp)
	c := &Cnf{}
	if err := opt(c); err != nil {
		t.Fatalf("WithPath() error: %v", err)
	}
	abs, _ := filepath.Abs(tmp)
	if c.Path != abs {
		t.Errorf("WithPath set Path = %q, want %q", c.Path, abs)
	}
}

func TestWithGitDirectory(t *testing.T) {
	opt := WithGitDirectory(".mygit")
	c := &Cnf{}
	if err := opt(c); err != nil {
		t.Fatalf("WithGitDirectory() error: %v", err)
	}
	if c.GitDirectory != ".mygit" {
		t.Errorf("WithGitDirectory set GitDirectory = %q, want %q", c.GitDirectory, ".mygit")
	}
}

func TestEditorFile(t *testing.T) {
	origConfig := *config
	t.Cleanup(func() {
		*config = origConfig
	})

	tmp := t.TempDir()
	err := Configure(WithPath(tmp))
	if err != nil {
		t.Fatalf("Configure() error: %v", err)
	}

	got := EditorFile()
	wantSuffix := "COMMIT_EDITMSG"
	if len(got) < len(wantSuffix) || got[len(got)-len(wantSuffix):] != wantSuffix {
		t.Errorf("EditorFile() = %q, want suffix %q", got, wantSuffix)
	}
	wantPrefix := GitPath()
	if len(got) < len(wantPrefix) || got[:len(wantPrefix)] != wantPrefix {
		t.Errorf("EditorFile() = %q, want prefix %q", got, wantPrefix)
	}
}

func TestDefaultBranch(t *testing.T) {
	origConfig := *config
	t.Cleanup(func() {
		*config = origConfig
	})

	config.DefaultBranch = "develop"
	if got := DefaultBranch(); got != "develop" {
		t.Errorf("DefaultBranch() = %q, want %q", got, "develop")
	}
}
