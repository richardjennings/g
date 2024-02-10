package index

import (
	"github.com/richardjennings/mygit/internal/mygit/config"
	"github.com/richardjennings/mygit/internal/mygit/gfs"
	"github.com/richardjennings/mygit/internal/mygit/objects"
)

// Status returns a FileSet containing all files from commit, index and working directory
// with the corresponding status.
func Status(idx *Index, commitSha []byte) (*gfs.FileSet, error) {
	var commitFiles []*gfs.File
	var err error
	if commitSha != nil {
		commitFiles, err = objects.CommittedFiles(commitSha)
		if err != nil {
			return nil, err
		}
	}
	files := gfs.NewFileSet(commitFiles)
	files.MergeFromIndex(gfs.NewFileSet(idx.Files()))
	workingDirectoryFiles, err := gfs.Ls(config.Path())
	if err != nil {
		return nil, err
	}
	files.MergeFromWD(gfs.NewFileSet(workingDirectoryFiles))
	return files, nil
}

// FsStatus returns a FileSet containing all files from the index and working directory
// with the corresponding status.
func FsStatus(path string) (*gfs.FileSet, error) {
	idx, err := ReadIndex()
	if err != nil {
		return nil, err
	}
	idxFiles := idx.Files()
	idxSet := gfs.NewFileSet(idxFiles)
	files, err := gfs.Ls(path)
	if err != nil {
		return nil, err
	}
	idxSet.MergeFromWD(gfs.NewFileSet(files))
	return idxSet, nil
}
