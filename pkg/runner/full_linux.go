package runner

import (
	"al.essio.dev/pkg/shellescape"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type MsysRunner struct {
	workdir string
	stdout  io.Writer
	stderr  io.Writer
}

func NewFullRunner(workdir string) *MsysRunner {
	return &MsysRunner{
		workdir: workdir,
		stdout:  os.Stdout,
		stderr:  os.Stderr,
	}
}

func (r *MsysRunner) GetDependencyString() string {
	return "core/fullrunner@latest"
}

func (r *MsysRunner) SetStdPipe(writer io.Writer) {
	r.stdout = writer
	r.stderr = writer
}

func (r *MsysRunner) CreateEnvironment() error {
	fmt.Print("Checking tar... ")
	if err := exec.Command("which", "tar").Run(); err != nil {
		fmt.Println("\t\tno")
		return err
	} else {
		fmt.Println("\t\tyes")
	}
	fmt.Print("Checking root...")
	if err := exec.Command("sudo", "-nv").Run(); err != nil {
		fmt.Println("\t\tno")
		return err
	} else {
		fmt.Println("\t\tyes")
	}
	fmt.Print("Extracting Arch rootfs...")
	if err := exec.Command(
		"tar", "--strip-components=1", "-xf",
		filepath.Join(r.workdir, "archlinux-bootstrap-x86_64.tar.zst"),
		"-C", r.workdir, "--numeric-owner",
	).Run(); err != nil {
		fmt.Println("\t\tfailed")
		return err
	} else {
		fmt.Println("\t\tdone")
	}
	os.Remove(filepath.Join(r.workdir, "archlinux-bootstrap-x86_64.tar.zst"))
	fmt.Print("Preparing pacman...")
	if err := r.Run("pacman-key --init", nil); err != nil {
		return err
	}
	if err := r.Run("pacman-key --populate", nil); err != nil {
		return err
	}
	if err := r.Run("echo 'Server = https://geo.mirror.pkgbuild.com/$repo/os/$arch' >> /etc/pacman.d/mirrorlist", nil); err != nil {
		return err
	}
	return r.Run("pacman -Sy --noconfirm", nil)
}

func (r *MsysRunner) Run(shellCommand string, env map[string]string) error {
	environ := os.Environ()
	if env == nil {
		env = map[string]string{}
	}
	for k, v := range env {
		environ = append(environ, k+"="+v)
	}
	cmd := exec.Command(filepath.Join(r.workdir, "bin", "arch-chroot"), r.workdir, "/bin/bash", "-lc", shellCommand)
	cmd.Dir = r.workdir
	cmd.Env = environ
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (r *MsysRunner) Copy(from string, to string, env map[string]string) error {
	return r.Run("cp "+shellescape.Quote(from)+" "+shellescape.Quote(to), env)
}
