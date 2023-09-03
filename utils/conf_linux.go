//go:build linux

package utils

import _ "embed"

//go:embed assets/7z.bin
var SevenZipExecutable []byte

const SevenZipName = "7z.bin"
