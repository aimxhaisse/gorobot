package gorobot

import (
	"os"
	"botapi"
	"fmt"
	"log"
	"time"
	"runtime"
)

// Writes the string to the log file, create/open the file if not yet open
func (robot *GoRobot) WriteLog(file string, msg string) {
	currentTime := time.LocalTime()
	if currentTime == nil {
		return
	}
	strTime := currentTime.String()
	fh, ok := robot.LogMap[file]
	if !ok {
		fh, _ = os.Open(file, os.O_WRONLY | os.O_CREAT | os.O_APPEND, 0666)
		if fh == nil {
			log.Printf("Warning: can't create file %s\n", file)
			return
		}
		robot.LogMap[file] = fh
	}
	fh.WriteString(fmt.Sprintf("%s - %s\n", strTime, msg))
}

// Logs a PRIVMSG event in logs/server/[user|channel].log
func (robot *GoRobot) LogEventPRIVMSG(ev *botapi.Event) {
	var file string

	if len(ev.Channel) > 0 {
		file = fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	} else {
		file = fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.User)
	}

	robot.WriteLog(file, fmt.Sprintf("%s: %s", ev.User, ev.Data))
}

// Logs a JOIN event in logs/server-channel.log
func (robot *GoRobot) LogEventJOIN(ev *botapi.Event) {
	file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	robot.WriteLog(file, fmt.Sprintf("%s has joined %s", ev.User, ev.Channel))
}

// Logs a PART event in logs/server-channel.log
func (robot *GoRobot) LogEventPART(ev *botapi.Event) {
	file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	robot.WriteLog(file, fmt.Sprintf("%s has left %s", ev.User, ev.Channel))
}

// Logs a KICK event in logs/server-channel.log
func (robot *GoRobot) LogEventKICK(ev *botapi.Event) {
	file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	robot.WriteLog(file, fmt.Sprintf("%s has been kicked from %s by %s", ev.Data, ev.Channel, ev.User))
}

// Main entry to log events
func (robot *GoRobot) LogEvent(ev *botapi.Event) {
	if robot.Config.Logs.Enable && robot.Config.Logs.RecordEvents {
		switch ev.Type {
		case botapi.E_PRIVMSG:
			robot.LogEventPRIVMSG(ev)
		case botapi.E_JOIN:
			robot.LogEventJOIN(ev)
		case botapi.E_PART:
			robot.LogEventPART(ev)
		case botapi.E_KICK:
			robot.LogEventKICK(ev)
		}
	}
}

// Periodically called to log some stats
func (robot *GoRobot) LogStatistics() {
	// Stats about Memory usage
	if robot.Config.Logs.Enable && robot.Config.Logs.RecordMemoryUsage {
		s := runtime.MemStats
		file := fmt.Sprintf("%s/memory.stats", robot.Config.Logs.Directory)
		robot.WriteLog(file, fmt.Sprintf("%d %d", s.Alloc, s.Sys))
	}
}
