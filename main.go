package main

import (
	"crond/schedule"
	"github.com/judwhite/go-svc/svc"
	"syscall"
)

type program struct {
	schedule *schedule.Schedule
}

func main() {
	prg := &program{}
	if err := svc.Run(prg, syscall.SIGINT, syscall.SIGTERM); err != nil {
		panic(err)
	}
}

func (p *program) Init(env svc.Environment) error {
	return nil
}

func (p *program) Start() error {
	opts := schedule.NewOptions()
	p.schedule = schedule.New(opts)
	p.schedule.Run()
	return nil
}

func (p *program) Stop() error {
	if p.schedule != nil {
		p.schedule.Exit()
	}
	return nil
}
