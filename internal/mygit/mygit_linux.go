//go:build linux

package mygit

import (
	"os"
	"syscall"
)

func setItemOsSpecificStat(info os.FileInfo, item *indexItem) {
	item.CTimeS = uint32(info.Sys().(*syscall.Stat_t).Ctim.Sec)
	item.CTimeN = uint32(info.Sys().(*syscall.Stat_t).Ctim.Nsec)
	item.MTimeS = uint32(info.Sys().(*syscall.Stat_t).Mtim.Sec)
	item.MTimeN = uint32(info.Sys().(*syscall.Stat_t).Mtim.Nsec)
}
