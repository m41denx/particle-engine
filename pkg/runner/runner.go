package runner

import "io"

type Runner interface {
	Copy(from string, to string) error
	Run(shellCommand string, env map[string]string) error
	SetStdPipe(writer io.Writer)
	CreateEnvironment() error
	GetDependencyString() string
}
