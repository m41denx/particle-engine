//go:build windows

package utils

import (
	_ "embed"
	"syscall"
)

//go:embed assets/7z.exe
var SevenZipExecutable []byte

const SevenZipName = "7z.exe"

const SymlinkPostfix = ".exe"

var Sysattr = &syscall.SysProcAttr{HideWindow: true}
