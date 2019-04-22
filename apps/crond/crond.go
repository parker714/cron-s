package main

import (
	"cron-s/crond"
	"github.com/judwhite/go-svc/svc"
	"log"
	"syscall"
)

type program struct {
	crond *crond.Crond
}

func main() {
	prg := &program{}
	if err := svc.Run(prg, syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}
}

func (p *program) Init(env svc.Environment) error {
	return nil
}

func (p *program) Start() error {
	opts := crond.NewOptions()
	p.crond = crond.New(opts)
	p.crond.Run()
	return nil
}

func (p *program) Stop() error {
	if p.crond != nil {
		p.crond.Exit()
	}
	return nil
}
