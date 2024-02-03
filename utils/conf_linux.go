//go:build linux

package utils

import _ "embed"
import "syscall"

//go:embed assets/7z.bin
var SevenZipExecutable []byte

const SevenZipName = "7z.bin"

const SymlinkPostfix = ""

var Sysattr = &syscall.SysProcAttr{}
