package g

import (
	"log"
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
			log.Fatalln(err)
		}
	}
	// set default main branch
	return UpdateHead(DefaultBranch())
}
