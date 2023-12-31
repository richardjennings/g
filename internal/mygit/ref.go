package mygit

import (
	"encoding/hex"
	"os"
	"path/filepath"
)

func (m *MyGit) updateHead(name string, sha []byte) error {
	path := filepath.Join(m.path, m.gitDirectory, RefsDirectory, "heads", name)
	return os.WriteFile(path, []byte(hex.EncodeToString(sha)), 0755)
}
