package mygit

import (
	"encoding/hex"
	"log"
	"time"
)

type (
/*
	commit struct {
		tree            string
		previousCommits []string
		author          string
		committer       string
		message         string
		date            time.Time
	}
*/
)

func (m *MyGit) Commit() error {
	author := "Richard Jennings <richardjennings@gmail.com>"
	committer := "Richard Jennings <richardjennings@gmail.com>"
	message := "test"

	// @todo this is pointed to by .git/HEAD
	currentBranch := "main"

	// get index
	index, err := m.readIndex()
	if err != nil {
		return err
	}
	files := index.idxFiles()

	// create trees of subtrees and blobs
	root := m.objectTree(files)
	sha, err := m.writeObjectTree(root)
	if err != nil {
		return err
	}

	// @todo check for a previous commit better
	var previousCommits [][]byte
	previousCommit, err := m.headSHA(currentBranch)
	if err != nil {
		return err
	}
	if previousCommit != nil {
		previousCommits = append(previousCommits, previousCommit)
	}

	commitSha, err := m.storeCommit(
		sha,
		previousCommits,
		author,
		time.Now(),
		committer,
		time.Now(),
		message,
	)
	if err != nil {
		return err
	}
	log.Println(hex.EncodeToString(commitSha))
	return m.updateHead(currentBranch, commitSha)
}
