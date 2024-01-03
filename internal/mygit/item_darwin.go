package mygit

import (
	"errors"
	"syscall"
)

func (wdf *wdFile) toIndexItem() (*indexItem, error) {
	if wdf.sha == nil {
		return nil, errors.New("missing sha from working directory file toIndexItem")
	}
	item := &indexItem{indexItemP: &indexItemP{}}
	item.CTimeS = uint32(wdf.finfo.Sys().(*syscall.Stat_t).Ctimespec.Sec)
	item.CTimeN = uint32(wdf.finfo.Sys().(*syscall.Stat_t).Ctimespec.Nsec)
	item.MTimeS = uint32(wdf.finfo.Sys().(*syscall.Stat_t).Mtimespec.Sec)
	item.MTimeN = uint32(wdf.finfo.Sys().(*syscall.Stat_t).Mtimespec.Nsec)
	item.Dev = uint32(wdf.finfo.Sys().(*syscall.Stat_t).Dev)
	item.Ino = uint32(wdf.finfo.Sys().(*syscall.Stat_t).Ino)
	if wdf.finfo.IsDir() {
		item.Mode = uint32(040000)
	} else {
		item.Mode = uint32(0100644)
	}
	item.Uid = wdf.finfo.Sys().(*syscall.Stat_t).Uid
	item.Gid = wdf.finfo.Sys().(*syscall.Stat_t).Gid
	item.Size = uint32(wdf.finfo.Size())
	copy(item.Sha[:], wdf.sha)
	nameLen := len(wdf.path)
	if nameLen < 0xFFF {
		item.Flags = uint16(len(wdf.path))
	} else {
		item.Flags = 0xFFF
	}
	item.Name = []byte(wdf.path)

	return item, nil
}
