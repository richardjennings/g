package g

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func UpdateHead(branch string) error {
	return os.WriteFile(GitHeadPath(), []byte(fmt.Sprintf("ref: refs/heads/%s\n", branch)), 0655)
}

// UpdateBranchHead updates the sha hash pointed to by a branch
func UpdateBranchHead(branch string, sha []byte) error {
	path := filepath.Join(RefsHeadsDirectory(), branch)
	return os.WriteFile(path, []byte(hex.EncodeToString(sha)+"\n"), 0755)
}

// HeadSHA returns the hash pointed to by a branch
func HeadSHA(currentBranch string) ([]byte, error) {
	path := filepath.Join(RefsHeadsDirectory(), currentBranch)
	bytes, err := os.ReadFile(path)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		// the default branch does not exist in refs/heads when there are no commits
		if currentBranch == DefaultBranchName {
			return nil, nil
		}
		return nil, fmt.Errorf("fatal: not a valid object name: '%s'", currentBranch)
	} else if err != nil {
		return nil, err
	} else if bytes == nil {
		return nil, fmt.Errorf("fatal: not a valid object name: '%s'", currentBranch)
	}
	return bytes[0:40], nil
}

// CurrentBranch returns the name of the current branch
func CurrentBranch() (string, error) {
	f, err := os.Open(GitHeadPath())
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

// ListBranches lists Git branches from refs/heads and info/refs
// It does not currently allow listing remote tracking branches
func ListBranches() ([]string, error) {
	var branches []string
	branchMap := make(map[string]struct{})

	// check for packed refs
	fh, err := os.Open(PackedRefsFile())
	defer func() {
		if fh != nil {
			_ = fh.Close()
		}
	}()
	if err == nil {
		// we have a packed refs file to parse into a list of branches
		scanner := bufio.NewScanner(fh)
		for scanner.Scan() {
			line := scanner.Bytes()
			// hash := line[0:40]
			path := string(line[41:])
			// path can have multiple prefixes
			// refs/heads/
			// refs (for stash)
			// refs/remotes/.../
			// for now just use refs/heads/
			if path, ok := strings.CutPrefix(path, RefsHeadPrefix()); ok {
				branchMap[path] = struct{}{}
			}
		}
	}

	f, err := os.ReadDir(RefsHeadsDirectory())
	if err != nil {
		return branches, err
	}
	for _, v := range f {
		if v.IsDir() {
			continue
		}
		branchMap[v.Name()] = struct{}{}
	}
	// return branches sorted alphabetically
	for k := range branchMap {
		branches = append(branches, k)
	}
	sort.Strings(branches)
	return branches, nil
}

func CreateBranch(name string) error {
	currentBranch, err := CurrentBranch()
	if err != nil {
		return err
	}
	head, err := HeadSHA(currentBranch)
	if err != nil {
		return err
	}

	sha := make([]byte, 20)
	if _, err := hex.Decode(sha, head); err != nil {
		return err
	}
	return UpdateBranchHead(name, sha)
}

func DeleteBranch(name string) error {
	return os.Remove(filepath.Join(RefsHeadsDirectory(), name))
}
