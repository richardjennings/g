package git

import (
	"errors"
	"fmt"
	"github.com/richardjennings/mygit/pkg/mygit/config"
	"github.com/richardjennings/mygit/pkg/mygit/index"
	"github.com/richardjennings/mygit/pkg/mygit/objects"
	"github.com/richardjennings/mygit/pkg/mygit/refs"
	"log"
	"os"
	"os/exec"
	"time"
)

// Commit writes a git commit object from the files in the index
func Commit(message []byte) ([]byte, error) {
	idx, err := index.ReadIndex()
	if err != nil {
		return nil, err
	}
	root := objects.ObjectTree(idx.Files())
	tree, err := root.WriteTree()
	if err != nil {
		return nil, err
	}
	// git has the --allow-empty flag which here defaults to true currently
	// @todo check for changes to be committed.
	previousCommits, err := refs.PreviousCommits()
	if err != nil {
		// @todo error types to check for e.g no previous commits as source of error
		return nil, err
	}
	commit := &objects.Commit{
		Tree:          tree,
		Parents:       previousCommits,
		Author:        fmt.Sprintf("%s <%s>", config.AuthorName(), config.AuthorEmail()),
		AuthoredTime:  time.Now(),
		Committer:     fmt.Sprintf("%s <%s>", config.CommitterName(), config.CommitterEmail()),
		CommittedTime: time.Now(),
	}
	if message != nil {
		commit.Message = message
	} else {
		// empty commit file
		if err := os.WriteFile(config.EditorFile(), []byte{}, 0600); err != nil {
			log.Fatalln(err)
		}
		ed, args := config.Editor()
		args = append(args, config.EditorFile())
		cmd := exec.Command(ed, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			log.Fatalln(err)
		}
		msg, err := os.ReadFile(args[0])
		if err != nil {
			log.Fatalln(msg)
		}
		commit.Message = msg
	}
	if len(commit.Message) == 0 {
		return nil, errors.New("Aborting commit due to empty commit message.")
	}
	return objects.WriteCommit(commit)
}
