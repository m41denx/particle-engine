package runner

import (
	"al.essio.dev/pkg/shellescape"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"slices"
)

type BusyboxRunner struct {
	workdir string
	stdout  io.Writer
	stderr  io.Writer
}

func NewThinRunner(workdir string) *BusyboxRunner {
	return &BusyboxRunner{
		workdir: workdir,
		stdout:  os.Stdout,
		stderr:  os.Stderr,
	}
}

func (r *BusyboxRunner) GetDependencyString() string {
	return "blank"
}

func (r *BusyboxRunner) SetStdPipe(writer io.Writer) {
	r.stdout = writer
	r.stderr = writer
}

func (r *BusyboxRunner) CreateEnvironment() error {
	j := func(dir ...string) string {
		return filepath.Join(slices.Concat([]string{r.workdir}, dir)...)
	}
	u, _ := user.Current()
	folders := []string{
		j("build"),
		j("dev"),
		j("etc"),
		j("home", u.Username),
		j("opt"),
		j("runnable"),
		j("tmp", "buildcache"),
		j("usr", "bin"),
		j("var"),
	}
	for _, folder := range folders {
		if err := os.MkdirAll(folder, 0755); err != nil {
			return err
		}
	}

	os.Setenv("PATH", os.Getenv("PATH")+":"+j("usr", "bin"))
	return nil
}

func (r *BusyboxRunner) Run(shellCommand string, env map[string]string) error {
	environ := os.Environ()
	if env == nil {
		env = map[string]string{}
	}
	for k, v := range env {
		environ = append(environ, k+"="+v)
	}
	cmd := exec.Command("/bin/bash", "-c", shellCommand)
	cmd.Dir = r.workdir
	cmd.Env = environ
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (r *BusyboxRunner) Copy(from string, to string, env map[string]string) error {
	return r.Run("cp "+shellescape.Quote(from)+" "+shellescape.Quote(to), env)
}
