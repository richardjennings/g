package git

import (
	"fmt"
	"github.com/richardjennings/mygit/pkg/mygit/config"
	"log"
	"os"
)

// Init initializes a git repository
func Init() error {
	path := config.GitPath()
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	for _, v := range []string{
		config.ObjectPath(),
		config.RefsDirectory(),
		config.RefsHeadsDirectory(),
	} {
		if err := os.MkdirAll(v, 0755); err != nil {
			log.Fatalln(err)
		}
	}
	// set default main branch
	return os.WriteFile(config.GitHeadPath(), []byte(fmt.Sprintf("ref: %s\n", config.Config.DefaultBranch)), 0644)
}
