package main

import "fmt"

type Command interface {
	Name() string
	Help() string
	Init(args []string) error
	Run() error
}

type arrayFlags []string

func (a *arrayFlags) String() string {
	return fmt.Sprintf("%v", *a)
}
func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func (a *arrayFlags) Get() interface{} {
	return *a
}
