package git

import (
	"fmt"
	"github.com/richardjennings/g"
	"path/filepath"
	"strings"
)

// Add adds one or more file paths to the Index.
func Add(paths ...string) error {
	idx, err := g.ReadIndex()
	if err != nil {
		return err
	}
	// get working directory files with idx status
	wdFiles, err := g.FsStatus(g.Path())
	if err != nil {
		return err
	}
	var updates []*g.File
	for _, p := range paths {
		if p == "." {
			// special case meaning add everything
			for _, v := range wdFiles.Files() {
				switch v.WdStatus {
				case g.WDUntracked, g.WDWorktreeChangedSinceIndex, g.WDDeletedInWorktree:
					updates = append(updates, v)
				}
			}
		} else {
			found := false
			for _, v := range wdFiles.Files() {
				if v.Path == p {
					switch v.WdStatus {
					case g.WDUntracked, g.WDWorktreeChangedSinceIndex, g.WDDeletedInWorktree:
						updates = append(updates, v)
					}
					found = true
					break
				}
			}
			if !found {
				// try directory @todo more efficient implementation
				for _, v := range wdFiles.Files() {
					if strings.HasPrefix(v.Path, p+string(filepath.Separator)) {
						switch v.WdStatus {
						case g.WDUntracked, g.WDWorktreeChangedSinceIndex, g.WDDeletedInWorktree:
							updates = append(updates, v)
						}
						found = true
					}
				}
			}

			if !found {
				return fmt.Errorf("fatal: pathspec '%s' did not match any files (directories not implemented yet)", p)
			}
		}
	}
	for _, v := range updates {
		switch v.WdStatus {
		case g.WDUntracked, g.WDWorktreeChangedSinceIndex:
			// add the file to the object store
			obj, err := g.WriteBlob(v.Path)
			if err != nil {
				return err
			}
			v.Sha, _ = g.NewSha(obj.Sha)
		}
		if err := idx.Add(v); err != nil {
			return err
		}
	}
	// once all files are added to idx struct, write it out
	return idx.Write()
}
