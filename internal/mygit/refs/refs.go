package refs

import (
	"bufio"
	"encoding/hex"
	"errors"
	"github.com/richardjennings/mygit/internal/mygit/config"
	"io/fs"
	"os"
	"path/filepath"
)

func UpdateHead(branch string, sha []byte) error {
	path := filepath.Join(config.RefsHeadsDirectory(), branch)
	return os.WriteFile(path, []byte(hex.EncodeToString(sha)+"\n"), 0755)
}

func HeadSHA(currentBranch string) ([]byte, error) {
	path := filepath.Join(config.RefsHeadsDirectory(), currentBranch)
	bytes, err := os.ReadFile(path)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return bytes[0:40], nil
}

func CurrentBranch() ([]byte, error) {
	f, err := os.Open(config.GitHeadPath())
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReader(f)
	b, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(b) < 12 {
		return nil, errors.New("invalid HEAD file, expected > 12 bytes")
	}

	return b[16 : len(b)-1], nil
}
