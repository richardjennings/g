package g

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// NotUpdated means that the modification time of the file in the
	// commit is the same as the one in the index.
	NotUpdated IndexStatus = iota

	// UpdatedInIndex means that the modification time of the file in the
	// index is newer than that of the commit
	UpdatedInIndex

	// TypeChangedInIndex mentioned in the Git docs but not implemented
	// here - file type changed (regular file, symbolic link or submodule)
	TypeChangedInIndex

	// AddedInIndex means that the file is in the index but not in the
	// commit.
	AddedInIndex

	// DeletedInIndex means that the file has been removed from the index
	DeletedInIndex

	// RenamedInIndex - @todo I do not understand how to implement this
	RenamedInIndex

	// CopiedInIndex - @todo not sure about this one either
	// - copied (if config option status.renames is set to "copies")
	CopiedInIndex

	// UntrackedInIndex means that the file is not in the Index
	UntrackedInIndex
)

const (
	// IndexAndWorkingTreeMatch means the file modification time in the
	// working directory is the same as in the index
	IndexAndWorkingTreeMatch WDStatus = iota

	// WorktreeChangedSinceIndex means the file in the working directory has
	// a newer modification time than the file in the index
	WorktreeChangedSinceIndex

	// TypeChangedInWorktreeSinceIndex is not implemented
	TypeChangedInWorktreeSinceIndex

	// DeletedInWorktree means that the file has been removed from the working
	// directory but exists in the commit
	DeletedInWorktree

	// RenamedInWorktree not implemented
	RenamedInWorktree

	// CopiedInWorktree not implemented
	CopiedInWorktree

	// Untracked means that the file is in the working directory but not in
	// the index or commit
	Untracked
)

type (
	FileStatus struct {
		path      string
		idxStatus IndexStatus
		wdStatus  WDStatus
		index     *fileInfo
		wd        *fileInfo
		commit    *fileInfo
	}
	fileInfo struct {
		Sha   Sha
		Finfo os.FileInfo
	}
	FfileSet struct {
		files []*FileStatus
		idx   map[string]*FileStatus
	}
	IndexStatus uint8
	WDStatus    uint8
	Finfo       struct {
		CTimeS uint32
		CTimeN uint32
		MTimeS uint32
		MTimeN uint32
		Dev    uint32
		Ino    uint32
		MMode  uint32
		Uid    uint32
		Gid    uint32
		SSize  uint32
		Sha    [20]byte
		NName  string
	}
	Mtime struct {
		Sec  uint32
		Nsec uint32
	}
)

func (is IndexStatus) StatusString() string {
	switch is {
	case NotUpdated:
		return " "
	case UpdatedInIndex:
		return "M"
	case TypeChangedInIndex:
		return "T"
	case AddedInIndex:
		return "A"
	case DeletedInIndex:
		return "D"
	case RenamedInIndex:
		return "R"
	case CopiedInIndex:
		return "C"
	case UntrackedInIndex:
		return "?"
	default:
		return ""
	}
}

func (wds WDStatus) StatusString() string {
	switch wds {
	case IndexAndWorkingTreeMatch:
		return " "
	case WorktreeChangedSinceIndex:
		return "M"
	case TypeChangedInWorktreeSinceIndex:
		return "T"
	case DeletedInWorktree:
		return "D"
	case RenamedInWorktree:
		return "R"
	case CopiedInWorktree:
		return "C"
	case Untracked:
		return "?"
	default:
		return ""
	}
}

// Ls recursively lists files in path that are not ignored
func Ls(path string) ([]*FileStatus, error) {
	var files []*FileStatus
	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// do not add ignored files
		if !IsIgnored(path, config.GitIgnore) {
			files = append(files, &FileStatus{
				path: strings.TrimPrefix(path, WorkingDirectory()),
				wd: &fileInfo{
					Finfo: info,
				},
			})
		}
		return nil
	}); err != nil {
		return files, err
	}
	return files, nil
}

func (fi *Finfo) Name() string {
	return fi.NName
}
func (fi *Finfo) Size() int64       { return int64(fi.SSize) }
func (fi *Finfo) Mode() os.FileMode { return 0 }
func (fi *Finfo) IsDir() bool       { return false }
func (fi *Finfo) Sys() any          { return nil }
func (fi *Finfo) ModTime() time.Time {
	return time.Unix(int64(fi.MTimeS), int64(fi.MTimeN))
}

func (f FileStatus) Path() string {
	return f.path
}
func (f FileStatus) IndexStatus() IndexStatus {
	return f.idxStatus
}
func (f FileStatus) WorkingDirectoryStatus() WDStatus {
	return f.wdStatus
}

func NewFfileSet(c []*FileStatus, i []*FileStatus, w []*FileStatus) (*FfileSet, error) {
	fs := &FfileSet{}
	fs.idx = make(map[string]*FileStatus)
	for _, v := range c {
		fs.idx[v.path] = v
		fs.files = append(fs.files, v)
	}
	fs.mergeFiles(i, 2)
	fs.mergeFiles(w, 3)
	return fs, fs.updateStatus()
}

func (f *FfileSet) updateStatus() error {
	for _, v := range f.files {
		// worktree status
		switch true {
		case v.index != nil && v.wd == nil:
			v.wdStatus = DeletedInWorktree
		case v.index != nil && v.wd != nil && v.index.Finfo.ModTime().Compare(v.wd.Finfo.ModTime()) == 0:
			v.wdStatus = IndexAndWorkingTreeMatch
		case v.index != nil && v.wd != nil && v.index.Finfo.ModTime().Compare(v.wd.Finfo.ModTime()) != 0:
			v.wdStatus = WorktreeChangedSinceIndex
		case v.index == nil && v.wd != nil:
			v.wdStatus = Untracked
		case v.commit != nil && v.index == nil && v.wd == nil:
			// in a commit but not in the index or wd
			v.wdStatus = IndexAndWorkingTreeMatch // because should flag as deleted in index status
		default:
			return errors.New("no matches for working tree status")
		}
		// index status
		switch true {
		case v.commit == nil && v.index == nil:
			v.idxStatus = UntrackedInIndex
		case v.commit != nil && v.index == nil:
			v.idxStatus = DeletedInIndex
		case v.commit != nil && v.index != nil && !v.index.Sha.Matches(v.commit.Sha):
			v.idxStatus = UpdatedInIndex
		case v.commit != nil && v.index != nil && v.index.Sha.Matches(v.commit.Sha):
			v.idxStatus = NotUpdated
		case v.commit == nil && v.index != nil:
			v.idxStatus = AddedInIndex
		default:
			return errors.New("no matches for index status")
		}
	}
	return nil
}

func (f *FfileSet) mergeFiles(files []*FileStatus, ciw int) {
	for _, v := range files {
		v := v
		ff, ok := f.idx[v.path]
		if ok {
			switch ciw {
			case 1:
				ff.commit = v.commit
			case 2:
				ff.index = v.index
			case 3:
				ff.wd = v.wd
			}
		} else {
			f.files = append(f.files, v)
			f.idx[v.path] = v
		}
	}
}

func (f *FfileSet) Files() []*FileStatus {
	return f.files
}

func (f *FfileSet) Contains(path string) (*FileStatus, bool) {
	v, ok := f.idx[path]
	return v, ok
}

func (wds WDStatus) String() string {
	switch wds {
	case IndexAndWorkingTreeMatch:
		return "IndexAndWorkingTreeMatch"
	case WorktreeChangedSinceIndex:
		return "WorktreeChangedSinceIndex"
	case TypeChangedInWorktreeSinceIndex:
		return "TypeChangedInWorktreeSinceIndex"
	case DeletedInWorktree:
		return "DeletedInWorktree"
	case RenamedInWorktree:
		return "RenamedInWorktree"
	case CopiedInWorktree:
		return "CopiedInWorktree"
	case Untracked:
		return "Untracked"
	default:
		return "UNKNOWN"
	}
}

func (is IndexStatus) String() string {
	switch is {
	case NotUpdated:
		return "NotUpdated"
	case UpdatedInIndex:
		return "UpdatedInIndex"
	case TypeChangedInIndex:
		return "TypeChangedInIndex"
	case AddedInIndex:
		return "AddedInIndex"
	case DeletedInIndex:
		return "DeletedInIndex"
	case RenamedInIndex:
		return "RenamedInIndex"
	case CopiedInIndex:
		return "CopiedInIndex"
	case UntrackedInIndex:
		return "UntrackedInIndex"
	default:
		return "UNKNOWN"
	}
}
