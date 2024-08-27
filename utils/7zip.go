package utils

import (
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
	if err := os.WriteFile(p, SevenZipExecutable, 0777); err != nil {
		panic(err)
	}
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
	cmd := exec.Command(s.binpath, args...)
	cmd.SysProcAttr = Sysattr
	cmd.Dir = s.workdir
	return cmd.Run()
}

func (s *SevenZip) Decompress(toDir string) error {
	args := []string{"x", s.archive, "-aoa", "-o" + toDir}
	args = append(args, s.args...)
	cmd := exec.Command(s.binpath, args...)
	cmd.SysProcAttr = Sysattr
	return cmd.Run()
}
