package conf

import (
	"time"
)

var (
	NodeId    string // The unique ID for this server across all time.
	HttpPort  string // The HTTP API port to listen on.
	Bootstrap bool   // This flag is used to control if a server is in "bootstrap" mode.
	Bind      string // The address that should be bound to for internal cluster communications.
	Join      string // Address of another agent to join upon starting up.
	DataDir   string // This flag provides a data directory for the agent to store state.

	DefaultScheduleTaskTick time.Duration
)

func init() {
	DefaultScheduleTaskTick = 1000 * time.Microsecond
}
