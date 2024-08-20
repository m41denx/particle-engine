package utils

import (
	"os"
)

type FileHold struct {
	Size       int64
	AccessTime int64
	Hash       string
}

func (fh *FileHold) Compare(path string, szOnly bool) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if fh.Size == info.Size() && szOnly {
		return true
	}
	md5Hash, err := CalcFileHash(path)
	if err != nil {
		return false
	}
	if md5Hash == fh.Hash {
		return true
	}

	return false
}

func NewFileHold(path string) *FileHold {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	fh := FileHold{
		Size:       info.Size(),
		AccessTime: info.ModTime().UnixNano(),
	}
	fh.Hash, _ = CalcFileHash(path)

	return &fh
}
