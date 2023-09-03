package utils

import (
	"github.com/bodgit/sevenzip"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

type SevenZip struct {
	binpath string
	archive string
	args    []string
	workdir string
}

func New7Zip() *SevenZip {
	p := path.Join(os.TempDir(), SevenZipName)
	os.WriteFile(p, SevenZipExecutable, 0750)
	return &SevenZip{binpath: p}
}

func (s *SevenZip) Session() *SevenZip {
	return &SevenZip{
		binpath: s.binpath,
		archive: s.archive,
		args:    s.args,
	}
}

func (s *SevenZip) OpenZip(archive string) *SevenZip {
	s.archive = archive
	return s
}

func (s *SevenZip) AddDirectory(dir string) *SevenZip {
	s.args = append(s.args, filepath.Join(dir, "*"))
	return s
}
func (s *SevenZip) WorkDir(dir string) *SevenZip {
	s.workdir = dir
	return s
}

func (s *SevenZip) Compress() error {
	args := []string{"a", s.archive}
	args = append(args, s.args...)
	//fmt.Println("\n", args)
	cmd := exec.Command(s.binpath, args...)
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	cmd.Dir = s.workdir
	return cmd.Run()
}

func (s *SevenZip) Decompress(toDir string) error {
	args := []string{"x", s.archive, "-aoa", "-o" + toDir}
	args = append(args, s.args...)
	cmd := exec.Command(s.binpath, args...)
	return cmd.Run()
}

func Un7zip(source, destination string) error {
	archive, err := sevenzip.OpenReader(source)
	if err != nil {
		return err
	}
	defer archive.Close()
	for _, file := range archive.Reader.File {
		reader, err := file.Open()
		if err != nil {
			return err
		}
		defer reader.Close()
		path := filepath.Join(destination, file.Name)
		// Remove file if it already exists; no problem if it doesn't; other cases can error out below
		_ = os.Remove(path)
		// Create a directory at path, including parents
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
		// If file is _supposed_ to be a directory, we're done
		if file.FileInfo().IsDir() {
			continue
		}
		// otherwise, remove that directory (_not_ including parents)
		err = os.Remove(path)
		if err != nil {
			return err
		}
		// and create the actual file.  This ensures that the parent directories exist!
		// An archive may have a single file with a nested path, rather than a file for each parent dir
		writer, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer writer.Close()
		_, err = io.Copy(writer, reader)
		if err != nil {
			return err
		}
	}
	return nil
}
