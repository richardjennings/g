package git

import (
	"errors"
	"fmt"
	"github.com/richardjennings/g"
	"log"
	"os"
	"os/exec"
	"time"
)

// Commit writes a git commit object from the files in the index
func Commit(message []byte) ([]byte, error) {
	idx, err := g.ReadIndex()
	if err != nil {
		return nil, err
	}
	root := g.ObjectTree(idx.Files())
	tree, err := root.WriteTree()
	if err != nil {
		return nil, err
	}
	// git has the --allow-empty flag which here defaults to true currently
	// @todo check for changes to be committed.
	previousCommits, err := g.PreviousCommits()
	if err != nil {
		// @todo error types to check for e.g no previous commits as source of error
		return nil, err
	}
	commit := &g.Commit{
		Tree:          tree,
		Parents:       previousCommits,
		Author:        fmt.Sprintf("%s <%s>", g.AuthorName(), g.AuthorEmail()),
		AuthoredTime:  time.Now(),
		Committer:     fmt.Sprintf("%s <%s>", g.CommitterName(), g.CommitterEmail()),
		CommittedTime: time.Now(),
	}
	if message != nil {
		commit.Message = message
	} else {
		// empty commit file
		if err := os.WriteFile(g.EditorFile(), []byte{}, 0600); err != nil {
			log.Fatalln(err)
		}
		ed, args := g.Editor()
		args = append(args, g.EditorFile())
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
	return g.WriteCommit(commit)
}
