package mygit

import (
	"log"
	"os"
	"path/filepath"
)

func (m *MyGit) Init() error {
	path := filepath.Join(m.path, m.gitDirectory)
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	for _, v := range []string{
		ObjectsDirectory,
		RefsDirectory,
		filepath.Join(RefsDirectory, RefsHeadsDirectory),
	} {
		if err := os.MkdirAll(filepath.Join(path, v), 0755); err != nil {
			log.Fatalln(err)
		}
	}
	// write an initially git HEAD
	os.WriteFile(filepath.Join(path, DefaultHeadFile), []byte("ref: refs/heads/main\n"), 0644)
	return nil
}
