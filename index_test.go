package g

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAdd_WDUntracked(t *testing.T) {
	assert.Nil(t, Configure())
	idx := NewIndex()
	err := idx.Add(&File{Path: "test_assets/test.file", Sha: Sha{}, wdStatus: WDUntracked})
	assert.Nil(t, err)
	files := idx.Files()
	assert.Equal(t, 1, len(files))
}
