package crond

import (
	"cron-s/internal/lg"
	"fmt"
)

func (c *Crond) logf(level lg.LogLevel, f string, args ...interface{}) {
	if c.Opts.LogLevel > level {
		return
	}
	c.Opts.Logger.Output(3, fmt.Sprintf(level.String()+": "+f, args...))
}
