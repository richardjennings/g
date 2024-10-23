package git

import (
	"fmt"
	"github.com/richardjennings/g"
	"log"
	"os"
)

// Init initializes a git repository
func Init() error {
	path := g.GitPath()
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	for _, v := range []string{
		g.ObjectPath(),
		g.RefsDirectory(),
		g.RefsHeadsDirectory(),
	} {
		if err := os.MkdirAll(v, 0755); err != nil {
			log.Fatalln(err)
		}
	}
	// set default main branch
	return os.WriteFile(g.GitHeadPath(), []byte(fmt.Sprintf("ref: %s\n", g.Config.DefaultBranch)), 0644)
}
