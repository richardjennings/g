package objects

import (
	"fmt"
	"github.com/richardjennings/mygit/internal/mygit/config"
	"path/filepath"
)

func (o *Object) WriteTree() ([]byte, error) {
	// resolve child tree Objects
	for i, v := range o.Objects {
		if v.Typ == ObjectTree {
			// if the tree only has blobs, write them and then
			// add the corresponding tree returning the Sha
			sha, err := v.WriteTree()
			if err != nil {
				return nil, err
			}
			o.Objects[i].Sha = sha
		}
	}
	// write a tree obj with the resolved children
	return o.writeTree()
}

func (o *Object) writeTree() ([]byte, error) {
	var content []byte
	var mode string
	for _, fo := range o.Objects {
		// @todo add executable support
		if fo.Typ == ObjectTree {
			mode = "40000"
		} else {
			mode = "100644"
		}
		// @todo replace base..
		content = append(content, []byte(fmt.Sprintf("%s %s%s%s", mode, filepath.Base(fo.Path), string(byte(0)), fo.Sha))...)
	}
	header := []byte(fmt.Sprintf("tree %d%s", len(content), string(byte(0))))
	return WriteObject(header, content, "", config.ObjectPath())
}
