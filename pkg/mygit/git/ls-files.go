package git

import "github.com/richardjennings/mygit/pkg/mygit/index"

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
