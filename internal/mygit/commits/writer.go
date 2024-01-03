package commits

import (
	"encoding/hex"
	"fmt"
	"github.com/richardjennings/mygit/internal/mygit/config"
	"github.com/richardjennings/mygit/internal/mygit/objects"
	"github.com/richardjennings/mygit/internal/mygit/refs"
)

func Write(c *Commit) ([]byte, error) {
	var parentCommits string
	for _, v := range c.Parents {
		parentCommits += fmt.Sprintf("parent %s\n", v)
	}
	content := []byte(fmt.Sprintf(
		"tree %s\n%sauthor %s %d +0000\ncommitter %s %d +0000\n\n%s",
		hex.EncodeToString(c.Tree),
		parentCommits,
		c.Author,
		c.AuthoredTime.Unix(),
		c.Committer,
		c.CommittedTime.Unix(),
		c.Message,
	))
	header := []byte(fmt.Sprintf("commit %d%s", len(content), string(byte(0))))
	sha, err := objects.WriteObject(header, content, "", config.ObjectPath())
	if err != nil {
		return nil, err
	}
	branch, err := refs.CurrentBranch()
	if err != nil {
		return nil, err
	}
	return sha, refs.UpdateHead(string(branch), sha)
}
