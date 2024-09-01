package runner

import (
	"al.essio.dev/pkg/shellescape"
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
	panic("MacOS Full runner is not supported yet")
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
	return r.Run("whoami", nil)
}

func (r *MsysRunner) Run(shellCommand string, env map[string]string) error {
	environ := os.Environ()
	if env == nil {
		env = map[string]string{}
	}
	for k, v := range env {
		environ = append(environ, k+"="+v)
	}
	cmd := exec.Command(filepath.Join(r.workdir, "usr", "bin", "bash.exe"), "-lc", shellCommand)
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
