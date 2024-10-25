package git

import (
	"github.com/richardjennings/g"
)

func SwitchBranch(name string) error {
	return g.SwitchToBranch(name)
}
