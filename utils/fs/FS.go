package fs

import "io"

type FS interface {
	GetFile(path string) ([]byte, error)
	PutFile(path string, data []byte) error
	PutFileStream(path string, data io.ReadCloser) error
	DeleteFile(path string) error
	DeleteFolder(path string) error
	DeleteList(objects []string) error
	DeleteFolderAsList(path string) error
	ListFolder(path string) ([]string, error)
}
