package crond

import (
	"flag"
	"time"
)

type options struct {
	nodeId    string // The unique ID for this server across all time.
	httpPort  string // The HTTP API port to listen on.
	bootstrap bool   // This flag is used to control if a server is in "bootstrap" mode.
	bind      string // The address that should be bound to for internal cluster communications.
	join      string // Address of another agent to join upon starting up.
	dataDir   string // This flag provides a data directory for the agent to store state.

	defaultScheduleTaskTick time.Duration
}

func NewOptions() *options {
	opts := &options{}

	flag.StringVar(&opts.nodeId, "node-id", "n0", "The unique ID for this server across all time.")
	flag.StringVar(&opts.httpPort, "http-port", ":7570", "The HTTP API port to listen on.")
	flag.BoolVar(&opts.bootstrap, "bootstrap", false, "This flag is used to control if a server is in 'bootstrap' mode.")
	flag.StringVar(&opts.bind, "bind", "127.0.0.1:8570", "The address that should be bound to for internal cluster communications.")
	flag.StringVar(&opts.join, "join", "", "Address of another agent to join upon starting up.")
	flag.StringVar(&opts.dataDir, "data-dir", "", "This flag provides a data directory for the agent to store state.")
	flag.Parse()

	opts.defaultScheduleTaskTick = 1000 * time.Microsecond

	return opts
}
