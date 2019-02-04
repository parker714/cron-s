package main

import (
	"cron-s/cronadmin"
	"github.com/judwhite/go-svc/svc"
	"log"
	"syscall"
)

type program struct {
	cronAdmin *cronadmin.CronAdmin
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
	opts := cronadmin.NewOptions()
	p.cronAdmin = cronadmin.New(opts)
	p.cronAdmin.Main()

	return nil
}

func (p *program) Stop() error {
	if p.cronAdmin != nil {
		p.cronAdmin.Exit()
	}
	return nil
}
