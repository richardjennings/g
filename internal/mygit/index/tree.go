package index

import (
	"github.com/richardjennings/mygit/internal/mygit/config"
	"github.com/richardjennings/mygit/internal/mygit/fs"
	"github.com/richardjennings/mygit/internal/mygit/objects"
	"path/filepath"
	"strings"
)

// ObjectTree creates a Tree Object with child Objects representing the files and
// paths in the provided files.
func ObjectTree(files []*fs.File) *objects.Object {
	root := &objects.Object{}
	var n *objects.Object  // current node
	var pn *objects.Object // previous node
	// mp holds a cache of file paths to objectTree nodes
	mp := make(map[string]*objects.Object)
	for _, v := range files {
		parts := strings.Split(strings.TrimPrefix(v.Path, config.WorkingDirectory()), string(filepath.Separator))
		if len(parts) == 1 {
			root.Objects = append(root.Objects, &objects.Object{Typ: objects.ObjectBlob, Path: v.Path, Sha: v.Sha})
			continue // top level file
		}
		pn = root
		for i, p := range parts {
			if i == len(parts)-1 {
				pn.Objects = append(pn.Objects, &objects.Object{Typ: objects.ObjectBlob, Path: v.Path, Sha: v.Sha})
				continue // leaf
			}
			// key for cached nodes
			key := strings.Join(parts[0:i+1], string(filepath.Separator))
			cached, ok := mp[key]
			if ok {
				n = cached
			} else {
				n = &objects.Object{Typ: objects.ObjectTree, Path: p}
				pn.Objects = append(pn.Objects, n)
				mp[key] = n
			}
			pn = n
		}
	}

	return root
}
