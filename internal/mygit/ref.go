package mygit

import (
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func (m *MyGit) updateHead(name string, sha []byte) error {
	path := filepath.Join(m.path, m.gitDirectory, RefsDirectory, "heads", name)
	return os.WriteFile(path, []byte(hex.EncodeToString(sha)+"\n"), 0755)
}

func (m *MyGit) headSHA(currentBranch string) ([]byte, error) {
	path := filepath.Join(m.path, m.gitDirectory, RefsDirectory, "heads", currentBranch)
	bytes, err := os.ReadFile(path)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return bytes[0:40], nil
}
