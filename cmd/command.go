package main

type Command interface {
	Name() string
	Help() string
	Init(args []string) error
	Run() error
}
