//go:build darwin

package g

import (
	"os"
	"syscall"
)

func setItemOsSpecificStat(info os.FileInfo, item *indexItem) {
	item.CTimeS = uint32(info.Sys().(*syscall.Stat_t).Ctimespec.Sec)
	item.CTimeN = uint32(info.Sys().(*syscall.Stat_t).Ctimespec.Nsec)
	item.MTimeS = uint32(info.Sys().(*syscall.Stat_t).Mtimespec.Sec)
	item.MTimeN = uint32(info.Sys().(*syscall.Stat_t).Mtimespec.Nsec)
}
