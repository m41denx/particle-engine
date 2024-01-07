package fs

import (
	"io"
	"os"
	"path/filepath"
)

type LocalFS struct {
	prefix string
}

func (l *LocalFS) GetFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(l.prefix, path))
}

func (l *LocalFS) PutFile(path string, data []byte) error {
	return os.WriteFile(filepath.Join(l.prefix, path), data, 0644)
}

func (l *LocalFS) PutFileStream(path string, data io.ReadCloser) error {
	f, err := os.OpenFile(filepath.Join(l.prefix, path), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, data)
	if err != nil {
		return err
	}
	return f.Close()
}

func (l *LocalFS) DeleteFile(path string) error {
	return os.Remove(filepath.Join(l.prefix, path))
}

func (l *LocalFS) DeleteFolder(path string) error {
	return os.RemoveAll(filepath.Join(l.prefix, path))
}

func (l *LocalFS) DeleteList(objects []string) error {
	for _, o := range objects {
		err := l.DeleteFile(o)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *LocalFS) DeleteFolderAsList(path string) error {
	list, err := l.ListFolder(path)
	if err != nil {
		return err
	}
	err = l.DeleteList(list)
	return err
}

func (l *LocalFS) ListFolder(path string) ([]string, error) {
	files, err := os.ReadDir(filepath.Join(l.prefix, path))
	if err != nil {
		return nil, err
	}
	var list []string
	for _, f := range files {
		list = append(list, f.Name())
	}
	return list, nil
}

func NewLocalFS() *LocalFS {
	home, _ := os.UserCacheDir()
	pc := filepath.Join(home, "particle_cache")
	return &LocalFS{
		prefix: filepath.Join(pc, "layers"),
	}
}
