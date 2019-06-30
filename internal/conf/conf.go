package conf

import (
	"time"
)

type Config struct {
	HttpPort      string
	Join          string
	WaitRenewTick time.Duration
	Raft          *Raft
}

type Raft struct {
	NodeId    string
	DataDir   string
	Bind      string
	Bootstrap bool
}

func New() *Config {
	return &Config{
		WaitRenewTick: time.Second * 3,
		Raft:          &Raft{},
	}
}
