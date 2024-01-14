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

// UpdateHead updates the sha hash pointed to by a branch
func UpdateHead(branch string, sha []byte) error {
	path := filepath.Join(config.RefsHeadsDirectory(), branch)
	return os.WriteFile(path, []byte(hex.EncodeToString(sha)+"\n"), 0755)
}

// HeadSHA returns the hash pointed to by a branch
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

// CurrentBranch returns the name of the current branch
func CurrentBranch() (string, error) {
	f, err := os.Open(config.GitHeadPath())
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReader(f)
	b, err := r.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	if len(b) < 12 {
		return "", errors.New("invalid HEAD file, expected > 12 bytes")
	}

	return string(b[16 : len(b)-1]), nil
}

// LastCommit return the last commit SHA on the current brand
func LastCommit() ([]byte, error) {
	currentBranch, err := CurrentBranch()
	if err != nil {
		return nil, err
	}
	return HeadSHA(currentBranch)
}

func PreviousCommits() ([][]byte, error) {
	previousCommit, err := LastCommit()
	if err != nil {
		return nil, err
	}
	if previousCommit != nil {
		return [][]byte{previousCommit}, nil
	}
	return nil, nil
}

func ListBranches() ([]string, error) {
	var branches []string
	f, err := os.ReadDir(config.RefsHeadsDirectory())
	if err != nil {
		return branches, err
	}
	for _, v := range f {
		if v.IsDir() {
			continue
		}
		branches = append(branches, v.Name())
	}
	return branches, nil
}
