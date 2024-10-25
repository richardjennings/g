package g

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	IndexNotUpdated IndexStatus = iota
	IndexUpdatedInIndex
	IndexTypeChangedInIndex
	IndexAddedInIndex
	IndexDeletedInIndex
	IndexRenamedInIndex
	IndexCopiedInIndex
	IndexUntracked
)

const (
	WDIndexAndWorkingTreeMatch WDStatus = iota
	WDWorktreeChangedSinceIndex
	WDTypeChangedInWorktreeSinceIndex
	WDDeletedInWorktree
	WDRenamedInWorktree
	WDCopiedInWorktree
	WDUntracked
)

type (
	File struct {
		Path      string
		idxStatus IndexStatus
		wdStatus  WDStatus
		Sha       Sha
		Finfo     os.FileInfo
	}
	Sha struct {
		set  bool
		hash [20]byte
	}
	IndexStatus uint8
	WDStatus    uint8
	FileSet     struct {
		files []*File
		idx   map[string]*File
	}
	Finfo struct {
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

// NewSha creates a Sha from either a binary or hex encoded byte slice
func NewSha(b []byte) (Sha, error) {
	if len(b) == 40 {
		s := Sha{set: true}
		_, _ = hex.Decode(s.hash[:], b)
		return s, nil
	}
	if len(b) == 20 {
		s := Sha{set: true}
		copy(s.hash[:], b)
		return s, nil
	}
	return Sha{}, fmt.Errorf("invalid sha %s", b)
}

func ShaFromHexString(s string) (Sha, error) {
	v, err := hex.DecodeString(s)
	if err != nil {
		return Sha{}, err
	}
	return NewSha(v)
}

func (s Sha) String() string {
	return s.AsHexString()
}

func (s Sha) IsSet() bool {
	return s.set
}

func (s Sha) AsHexString() string {
	return hex.EncodeToString(s.hash[:])
}

func (s Sha) AsHexBytes() []byte {
	b := make([]byte, 40)
	hex.Encode(b, s.hash[:])
	return b
}

func (s Sha) AsArray() [20]byte {
	var r [20]byte
	copy(r[:], s.hash[:])
	return r
}

// AsByteSlice returns a Sha as a byte slice
func (s Sha) AsByteSlice() []byte {
	return s.hash[:]
}

func (is IndexStatus) String() string {
	switch is {
	case IndexNotUpdated:
		return " "
	case IndexUpdatedInIndex:
		return "M"
	case IndexTypeChangedInIndex:
		return "T"
	case IndexAddedInIndex:
		return "A"
	case IndexDeletedInIndex:
		return "D"
	case IndexRenamedInIndex:
		return "R"
	case IndexCopiedInIndex:
		return "C"
	case IndexUntracked:
		return "?"
	default:
		return ""
	}
}

func (wds WDStatus) String() string {
	switch wds {
	case WDIndexAndWorkingTreeMatch:
		return " "
	case WDWorktreeChangedSinceIndex:
		return "M"
	case WDTypeChangedInWorktreeSinceIndex:
		return "T"
	case WDDeletedInWorktree:
		return "D"
	case WDRenamedInWorktree:
		return "R"
	case WDCopiedInWorktree:
		return "C"
	case WDUntracked:
		return "?"
	default:
		return ""
	}
}

// Ls recursively lists files in path
func Ls(path string) ([]*File, error) {
	var files []*File
	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// do not add ignored files
		if !IsIgnored(path) {
			files = append(files, &File{
				Path:  strings.TrimPrefix(path, WorkingDirectory()),
				Finfo: info,
			})
		}
		return nil
	}); err != nil {
		return files, err
	}
	return files, nil
}

func NewFileSet(files []*File) *FileSet {
	fs := &FileSet{files: files}
	fs.idx = make(map[string]*File)
	for _, v := range files {
		fs.idx[v.Path] = v
	}
	return fs
}

func (fs *FileSet) Merge(fss *FileSet) {
	for _, v := range fss.idx {
		if _, ok := fs.idx[v.Path]; ok {
			if fs.idx[v.Path].idxStatus == IndexNotUpdated {
				fs.idx[v.Path].idxStatus = v.idxStatus
			}
			if fs.idx[v.Path].wdStatus == WDIndexAndWorkingTreeMatch {
				fs.idx[v.Path].wdStatus = v.wdStatus
			}
		} else {
			fs.files = append(fs.files, v)
			fs.idx[v.Path] = v
		}
	}
}

func (fs *FileSet) MergeFromIndex(fss *FileSet) {
	// add files from index to set, updating the status as relevant
	for _, v := range fss.files {
		if _, ok := fs.idx[v.Path]; !ok {
			// in index but not in commit files
			fs.files = append(fs.files, v)
			fs.idx[v.Path] = v
			v.idxStatus = IndexAddedInIndex
			continue
		}
		fs.idx[v.Path].Finfo = v.Finfo
		if !bytes.Equal(v.Sha.AsByteSlice(), fs.idx[v.Path].Sha.AsByteSlice()) {
			fs.idx[v.Path].idxStatus = IndexUpdatedInIndex
			continue
		}
	}
	for _, v := range fs.files {
		if _, ok := fss.idx[v.Path]; !ok {
			// file exists in commit but not in index
			v.idxStatus = IndexDeletedInIndex
		}
	}
}

func (fs *FileSet) MergeFromWD(fss *FileSet) {
	// add files from working directory to set, updating the status as relevant
	for _, v := range fss.files {
		if _, ok := fs.idx[v.Path]; !ok {
			// in working directory but not in index or commit files
			fs.files = append(fs.files, v)
			fs.idx[v.Path] = v
			v.wdStatus = WDUntracked
			v.idxStatus = IndexUntracked
			continue
		}

		if fs.idx[v.Path].Finfo == nil {
			// this is a commit file and not in the index
			// @todo should this be able to happen ?
			fs.idx[v.Path].wdStatus = WDUntracked
			fs.idx[v.Path].idxStatus = IndexUntracked
		} else {
			if v.Finfo.ModTime() != fs.idx[v.Path].Finfo.ModTime() {
				fs.idx[v.Path].wdStatus = WDWorktreeChangedSinceIndex
				fs.idx[v.Path].Finfo = v.Finfo
				// flag that the object needs to be indexed
				// perhaps index add should be smarter instead ?
				fs.idx[v.Path].Sha = Sha{}
				continue
			}

		}
	}
	for _, v := range fs.files {
		if _, ok := fss.idx[v.Path]; !ok {
			// file exists in commit but not in index
			v.wdStatus = WDDeletedInWorktree
		}
	}
}

func (fs *FileSet) Add(file *File) {
	fs.idx[file.Path] = file
	fs.files = append(fs.files, file)
}

// Compliment returns files in s that are not in fs
func (fs *FileSet) Compliment(s *FileSet) *FileSet {
	r := NewFileSet(nil)
	for k, v := range s.idx {
		if _, ok := fs.idx[k]; !ok {
			r.Add(v)
		}
	}
	return r
}

func (fs *FileSet) Contains(path string) (*File, bool) {
	v, ok := fs.idx[path]
	return v, ok
}

func (fs *FileSet) Files() []*File {
	return fs.files
}

// os.FileInfo interface

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

// Init initializes a git repository
func Init() error {
	path := GitPath()
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	for _, v := range []string{
		ObjectPath(),
		RefsDirectory(),
		RefsHeadsDirectory(),
	} {
		if err := os.MkdirAll(v, 0755); err != nil {
			log.Fatalln(err)
		}
	}
	// set default main branch
	return os.WriteFile(GitHeadPath(), []byte(fmt.Sprintf("ref: %s\n", filepath.Join(RefsHeadPrefix(), DefaultBranch()))), 0644)
}

func (f File) IndexStatus() IndexStatus {
	return f.idxStatus
}

func (f File) WorkingDirectoryStatus() WDStatus {
	return f.wdStatus
}

// SwitchToBranch updates the repository content to match that of a specified branch name
// or returns an error when it is not safe to do so. This should likely be cahnged to
// SwitchToCommit in the future to handle the broader use-case.
func SwitchToBranch(name string) error {

	// get commit sha
	commitSha, err := HeadSHA(name)
	if err != nil {
		return err
	}

	if !commitSha.IsSet() {
		return fmt.Errorf("fatal: invalid reference: %s", name)
	}

	// index
	idx, err := ReadIndex()
	if err != nil {
		return err
	}

	currentCommit, err := LastCommit()
	if err != nil {
		// @todo error types to check for e.g no previous commits as source of error
		return err
	}

	//

	currentStatus, err := Status(idx, currentCommit)
	if err != nil {
		return err
	}

	// get commit files
	commitFiles, err := CommittedFiles(commitSha)
	if err != nil {
		return err
	}

	commitSet := NewFileSet(commitFiles)

	var errorWdFiles []*File
	var errorIdxFiles []*File
	var deleteFiles []*File

	for _, v := range currentStatus.Files() {
		if v.IndexStatus() == IndexUpdatedInIndex {
			errorIdxFiles = append(errorIdxFiles, v)
			continue
		}
		if _, ok := commitSet.Contains(v.Path); ok {
			if v.WorkingDirectoryStatus() == WDUntracked {
				errorWdFiles = append(errorWdFiles, v)
				continue
			}
		} else {
			// should be deleted
			deleteFiles = append(deleteFiles, v)
		}
	}
	var errMsg = ""
	if len(errorIdxFiles) > 0 {
		filestr := ""
		for _, v := range errorIdxFiles {
			filestr += fmt.Sprintf("\t%s\n", v.Path)
		}
		errMsg = fmt.Sprintf("error: The following untracked working tree files would be overwritten by checkout:\n %sPlease move or remove them before you switch branches.\nAborting", filestr)
	}
	if len(errorWdFiles) > 0 {
		filestr := ""
		for _, v := range errorWdFiles {
			filestr += fmt.Sprintf("\t%s\n", v.Path)
		}
		if errMsg != "" {
			errMsg += "\n"
		}
		errMsg += fmt.Sprintf("error: The following untracked working tree files would be overwritten by checkout:\n %sPlease move or remove them before you switch branches.\nAborting", filestr)
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}

	for _, v := range deleteFiles {
		if err := os.Remove(filepath.Join(Path(), v.Path)); err != nil {
			return err
		}
	}

	idx = NewIndex()

	for _, v := range commitFiles {
		obj, err := ReadObject(v.Sha)
		if err != nil {
			return err
		}
		r, err := obj.ReadCloser()
		if err != nil {
			return err
		}
		buf := make([]byte, obj.HeaderLength)
		if _, err := r.Read(buf); err != nil {
			return err
		}
		f, err := os.OpenFile(filepath.Join(Path(), v.Path), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0655)
		if err != nil {
			return err
		}

		if _, err := io.Copy(f, r); err != nil {
			return err
		}
		if err := r.Close(); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}

		v.wdStatus = WDUntracked
		if err := idx.Add(v); err != nil {
			return err
		}
	}

	if err := idx.Write(); err != nil {
		return err
	}

	// update HEAD
	if err := UpdateHead(name); err != nil {
		return err
	}

	return nil
}
