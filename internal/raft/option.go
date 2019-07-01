package raft

// Option is the config struct
type Option struct {
	NodeID    string
	DataDir   string
	Bind      string
	Bootstrap bool
}

// NewOption returns raft config instance
func NewOption() *Option {
	return &Option{}
}
