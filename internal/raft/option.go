package raft

type Option struct {
	NodeId    string
	DataDir   string
	Bind      string
	Bootstrap bool
}

func NewOption() *Option {
	return &Option{}
}
