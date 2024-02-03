//go:build darwin

package utils

import _ "embed"
import "syscall"

//go:embed assets/7zd.bin
var SevenZipExecutable []byte

const SevenZipName = "7zd.bin"

const SymlinkPostfix = ""

var Sysattr = &syscall.SysProcAttr{}
