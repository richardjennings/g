package mygit

import (
	"encoding/hex"
	"log"
)

func (m *MyGit) Commit() error {
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
	commitSha, err := m.storeCommit(sha, [][]byte{}, "Richard Jennings <richardjennings@gmail.com>", "Richard Jennings <richardjennings@gmail.com>", "test")
	if err != nil {
		return err
	}
	log.Println(hex.EncodeToString(commitSha))
	return m.updateHead("main", commitSha)
}
