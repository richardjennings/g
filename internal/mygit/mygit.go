package mygit

import (
	"errors"
	"fmt"
	"github.com/richardjennings/mygit/internal/mygit/config"
	"github.com/richardjennings/mygit/internal/mygit/fs"
	"github.com/richardjennings/mygit/internal/mygit/index"
	"github.com/richardjennings/mygit/internal/mygit/objects"
	"github.com/richardjennings/mygit/internal/mygit/refs"
	"io"
	"log"
	"os"
	"time"
)

// Init initializes a git repository
func Init() error {
	path := config.GitPath()
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	for _, v := range []string{
		config.ObjectPath(),
		config.RefsDirectory(),
		config.RefsHeadsDirectory(),
	} {
		if err := os.MkdirAll(v, 0755); err != nil {
			log.Fatalln(err)
		}
	}
	// set default main branch
	return os.WriteFile(config.GitHeadPath(), []byte(fmt.Sprintf("ref: %s\n", config.Config.DefaultBranch)), 0644)
}

// Log prints out the commit log for the current branch
func Log() error {
	branch, err := refs.CurrentBranch()
	if err != nil {
		return err
	}
	commitSha, err := refs.HeadSHA(branch)
	if err != nil {
		return err
	}
	commit, err := objects.ReadCommit(commitSha)
	if err != nil {
		return err
	}
	fmt.Printf("tree: %s\n", string(commit.Tree))
	for _, v := range commit.Parents {
		fmt.Printf("parent: %s\n", string(v))
	}
	return nil
}

// Add adds one or more file paths to the Index.
func Add(paths ...string) error {
	idx, err := index.ReadIndex()
	if err != nil {
		return err
	}
	// get working directory files with idx status
	wdFiles, err := index.FsStatus(config.Path())
	if err != nil {
		return err
	}
	var updates []*fs.File
	for _, p := range paths {
		if p == "." {
			// special case meaning add everything
			for _, v := range wdFiles {
				switch v.Status {
				case fs.StatusUntracked, fs.StatusModified, fs.StatusDeleted:
					updates = append(updates, v)
				}
			}
		} else {
			// @todo add support for paths other than just '.'
			return errors.New("only supports '.' currently ")
		}
	}
	for _, v := range updates {
		switch v.Status {
		case fs.StatusUntracked, fs.StatusModified:
			// add the file to the object store
			obj, err := objects.WriteBlob(v.Path)
			if err != nil {
				return err
			}
			v.Sha = obj.Sha
		}
		if err := idx.Add(v); err != nil {
			return err
		}
	}
	// once all files are added to idx struct, write it out
	return idx.Write()
}

// LsFiles returns a list of files in the index
func LsFiles() ([]string, error) {
	idx, err := index.ReadIndex()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, v := range idx.Files() {
		files = append(files, v.Path)
	}
	return files, nil
}

// Commit writes a git commit object from the files in the index
func Commit() ([]byte, error) {
	idx, err := index.ReadIndex()
	if err != nil {
		return nil, err
	}
	root := index.ObjectTree(idx.Files())
	tree, err := root.WriteTree()
	if err != nil {
		return nil, err
	}
	// git has the --allow-empty flag which here defaults to true currently
	// @todo check for changes to be committed.
	previousCommits, err := refs.PreviousCommits()
	if err != nil {
		return nil, err
	}
	return objects.WriteCommit(
		&objects.Commit{
			Tree:          tree,
			Parents:       previousCommits,
			Author:        "Richard Jennings <richardjennings@gmail.com>",
			AuthoredTime:  time.Now(),
			Committer:     "Richard Jennings <richardjennings@gmail.com>",
			CommittedTime: time.Now(),
			Message:       "test",
		},
	)
}

// Status currently displays the file statuses comparing the working directory
// to the index and the index to the last commit (if any).
func Status(o io.Writer) error {
	var err error
	// index
	idx, err := index.ReadIndex()
	if err != nil {
		return err
	}
	commitSha, err := refs.LastCommit()
	if err != nil {
		return err
	}
	files, err := idx.CommitStatus(commitSha)
	if err != nil {
		return err
	}
	for _, v := range files {
		if v.Status == fs.StatusUnchanged {
			continue
		}
		if _, err := fmt.Fprintf(o, "%s  %s\n", v.Status, v.Path); err != nil {
			return err
		}
	}

	// working directory
	files, err = index.FsStatus(config.Path())
	if err != nil {
		return err
	}
	for _, v := range files {
		if v.Status == fs.StatusUnchanged {
			continue
		}
		if _, err := fmt.Fprintf(o, " %s %s\n", v.Status, v.Path); err != nil {
			return err
		}
	}
	return nil
}
