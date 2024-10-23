package git

import (
	"errors"
	"fmt"
	"github.com/richardjennings/g"
	"io"
	"os"
	"path/filepath"
)

func SwitchBranch(name string) error {

	// index
	idx, err := g.ReadIndex()
	if err != nil {
		return err
	}

	// get commit sha
	commitSha, err := g.HeadSHA(name)
	if err != nil {
		return err
	}

	if commitSha == nil {
		return fmt.Errorf("fatal: invalid reference: %s", name)
	}

	currentCommit, err := g.LastCommit()
	if err != nil {
		// @todo error types to check for e.g no previous commits as source of error
		return err
	}

	currentStatus, err := g.Status(idx, currentCommit)
	if err != nil {
		return err
	}

	// get commit files
	commitFiles, err := g.CommittedFiles(commitSha)
	if err != nil {
		return err
	}

	commitSet := g.NewFileSet(commitFiles)

	var errorWdFiles []*g.File
	var errorIdxFiles []*g.File
	var deleteFiles []*g.File

	for _, v := range currentStatus.Files() {
		if v.IdxStatus == g.IndexUpdatedInIndex {
			errorIdxFiles = append(errorIdxFiles, v)
			continue
		}
		if _, ok := commitSet.Contains(v.Path); ok {
			if v.WdStatus == g.WDUntracked {
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
		if err := os.Remove(filepath.Join(g.Path(), v.Path)); err != nil {
			return err
		}
	}

	idx = g.NewIndex()

	for _, v := range commitFiles {
		obj, err := g.ReadObject(v.Sha.AsHexBytes())
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
		f, err := os.OpenFile(filepath.Join(g.Path(), v.Path), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0655)
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
		v.WdStatus = g.WDUntracked
		if err := idx.Add(v); err != nil {
			return err
		}
	}

	if err := idx.Write(); err != nil {
		return err
	}

	// update HEAD
	if err := g.UpdateHead(name); err != nil {
		return err
	}

	return nil

}
