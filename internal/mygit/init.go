package mygit

import (
	"log"
	"os"
	"path/filepath"
)

func (m *MyGit) Init(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	path = filepath.Join(path, m.gitDirectory)
	if err := os.MkdirAll(path, 0744); err != nil {
		return err
	}
	for _, v := range []string{
		ObjectsDirectory,
		RefsDirectory,
	} {
		if err := os.MkdirAll(filepath.Join(path, v), 0755); err != nil {
			log.Fatalln(err)
		}
	}
	return nil
}
