package main

type Runner interface {
	Run() error
	Name() string
}

type Service interface {
	Router
	Runner
}
