package mygit

import (
	"encoding/hex"
	"log"
)

func (m *MyGit) Commit() error {
	currentBranch := "main"
	files, err := m.files()
	if err != nil {
		return err
	}
	var fos []*fileObject
	for _, v := range files {
		fo, err := m.storeBlob(v)
		if err != nil {
			return err
		}
		fos = append(fos, fo)
	}
	sha, err := m.storeTree(fos)
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
		"Richard Jennings <richardjennings@gmail.com>",
		"Richard Jennings <richardjennings@gmail.com>",
		"test",
	)
	if err != nil {
		return err
	}
	log.Println(hex.EncodeToString(commitSha))
	return m.updateHead(currentBranch, commitSha)
}
