package index

import (
	"github.com/richardjennings/mygit/pkg/gfs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAdd_WDUntracked(t *testing.T) {
	idx := NewIndex()
	err := idx.Add(&gfs.File{Path: "test_assets/test.file", Sha: &gfs.Sha{}, WdStatus: gfs.WDUntracked})
	assert.Nil(t, err)
	files := idx.Files()
	assert.Equal(t, 1, len(files))
}
