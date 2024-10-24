package g

import (
	"bufio"
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
func UpdateBranchHead(branch string, sha Sha) error {
	path := filepath.Join(RefsHeadsDirectory(), branch)
	return os.WriteFile(path, []byte(sha.AsHexString()+"\n"), 0755)
}

// HeadSHA returns the hash pointed to by a branch
func HeadSHA(currentBranch string) (Sha, error) {
	path := filepath.Join(RefsHeadsDirectory(), currentBranch)
	bytes, err := os.ReadFile(path)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		// the branch does not exist in refs/heads when there are no commits
		// lets check packed-refs
		branchMap, err := packedrefs()
		if err != nil {
			return Sha{}, err
		}
		if v, ok := branchMap[currentBranch]; ok {
			return v, nil
		}
		return Sha{}, nil
	} else if err != nil {
		return Sha{}, err
	} else if bytes == nil {
		return Sha{}, fmt.Errorf("fatal: not a valid object name: '%s'", currentBranch)
	}
	sha, err := NewSha(bytes[0:40])
	return sha, err
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
func LastCommit() (Sha, error) {
	currentBranch, err := CurrentBranch()
	if err != nil {
		return Sha{}, err
	}
	sha, err := HeadSHA(currentBranch)
	if err != nil {
		return Sha{}, err
	}
	return sha, nil
}

func PreviousCommits() ([]Sha, error) {
	previousCommit, err := LastCommit()
	if err != nil {
		return nil, err
	}
	if previousCommit.IsSet() {
		return []Sha{previousCommit}, nil
	}
	return nil, nil
}

// ListBranches lists Git branches from refs/heads and info/refs
// It does not currently allow listing remote tracking branches
func ListBranches() ([]string, error) {
	var branches []string
	branchMap := make(map[string]struct{})

	// check for packed refs
	packed, err := packedrefs()
	if err != nil {
		return nil, err
	}
	for k := range packed {
		branchMap[k] = struct{}{}
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

	return UpdateBranchHead(name, head)
}

func DeleteBranch(name string) error {
	return os.Remove(filepath.Join(RefsHeadsDirectory(), name))
}

func packedrefs() (map[string]Sha, error) {
	branchMap := make(map[string]Sha)

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
			hash := line[0:40]
			path := string(line[41:])
			// path can have multiple prefixes
			// refs/heads/
			// refs (for stash)
			// refs/remotes/.../
			// for now just use refs/heads/
			if path, ok := strings.CutPrefix(path, RefsHeadPrefix()); ok {
				branchMap[path], err = NewSha(hash)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return branchMap, nil
}
