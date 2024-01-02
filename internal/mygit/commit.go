package mygit

import (
	"encoding/hex"
	"log"
)

func (m *MyGit) Commit() error {
	currentBranch := "main"
	author := "Richard Jennings <richardjennings@gmail.com>"
	committer := "Richard Jennings <richardjennings@gmail.com>"
	message := "test"

	// get index
	index, err := m.readIndex()
	if err != nil {
		return err
	}
	files := index.fileNames()
	//files, err := m.files()
	//if err != nil {
	//	return err
	//}

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
		committer,
		message,
	)
	if err != nil {
		return err
	}
	log.Println(hex.EncodeToString(commitSha))
	return m.updateHead(currentBranch, commitSha)
}
