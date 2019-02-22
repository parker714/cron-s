// short for "log"
package lg

import (
	"fmt"
	"log"
	"os"
)

type Lg struct {
	AppName     string
	AppLogLevel LogLevel
	logger      *log.Logger
}

func New(AppName string, LogLevel LogLevel) *Lg {
	lg := &Lg{
		AppName:     AppName,
		AppLogLevel: LogLevel,
	}

	lg.logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	return lg
}

func (l *Lg) Logf(level LogLevel, f string, args ...interface{}) {
	if l.AppLogLevel > level {
		return
	}
	err := l.logger.Output(3, fmt.Sprintf(level.String()+" | "+l.AppName+": "+f, args...))
	if err != nil {
		panic(err)
	}
}
