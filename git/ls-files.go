package git

import "github.com/richardjennings/g"

// LsFiles returns a list of files in the index
func LsFiles() ([]string, error) {
	idx, err := g.ReadIndex()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, v := range idx.Files() {
		files = append(files, v.Path)
	}
	return files, nil
}
