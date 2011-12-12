package main

import (
	"github.com/aimxhaisse/gorobot/api"
	"os"
	"fmt"
	"log"
	"time"
	"runtime"
)

func (robot *GoRobot) writeLog(file string, msg string) {
	currentTime := time.LocalTime()
	if currentTime == nil {
		return
	}
	strTime := currentTime.String()
	fh, ok := robot.LogMap[file]
	if !ok {
		fh, _ = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if fh == nil {
			log.Printf("Warning: can't create file %s\n", file)
			return
		}
		robot.LogMap[file] = fh
	}
	fh.WriteString(fmt.Sprintf("%s - %s\n", strTime, msg))
}

func (robot *GoRobot) logEventPRIVMSG(ev *api.Event) {
	var file string

	if len(ev.Channel) > 0 {
		file = fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	} else {
		file = fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.User)
	}

	robot.writeLog(file, fmt.Sprintf("%s: %s", ev.User, ev.Data))
}

func (robot *GoRobot) logEventJOIN(ev *api.Event) {
	file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	robot.writeLog(file, fmt.Sprintf("%s has joined %s", ev.User, ev.Channel))
}

func (robot *GoRobot) logEventPART(ev *api.Event) {
	file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	robot.writeLog(file, fmt.Sprintf("%s has left %s", ev.User, ev.Channel))
}

func (robot *GoRobot) logEventKICK(ev *api.Event) {
	file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	robot.writeLog(file, fmt.Sprintf("%s has been kicked from %s by %s", ev.Data, ev.Channel, ev.User))
}

// LogEvent logs events
func (robot *GoRobot) LogEvent(ev *api.Event) {
	if robot.Config.Logs.Enable && robot.Config.Logs.RecordEvents {
		switch ev.Type {
		case api.E_PRIVMSG:
			robot.logEventPRIVMSG(ev)
		case api.E_JOIN:
			robot.logEventJOIN(ev)
		case api.E_PART:
			robot.logEventPART(ev)
		case api.E_KICK:
			robot.logEventKICK(ev)
		}
	}
}

// LogStatistics is periodically called to log statistics about the memory usage of the IRC robot
func (robot *GoRobot) LogStatistics() {
	if robot.Config.Logs.Enable && robot.Config.Logs.RecordMemoryUsage {
		s := runtime.MemStats
		file := fmt.Sprintf("%s/memory.stats", robot.Config.Logs.Directory)
		robot.writeLog(file, fmt.Sprintf("%d %d", s.Alloc, s.Sys))
	}
}
