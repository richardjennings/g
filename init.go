package g

import (
	"fmt"
	"os"
)

// Init initializes a git repository
func Init() error {
	path := GitPath()
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	for _, v := range []string{
		ObjectPath(),
		RefsDirectory(),
		RefsHeadsDirectory(),
	} {
		if err := os.MkdirAll(v, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", v, err)
		}
	}
	// set default main branch
	return UpdateHead(DefaultBranch())
}
