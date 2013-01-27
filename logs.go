package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Writes the string to the log file, creates the log file if it doesn't exists
func (robot *Bot) writeLog(file string, msg string) {
	currentTime := time.Now()
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

func (robot *Bot) logEventPRIVMSG(ev *Event) {
	var file string

	if len(ev.Channel) > 0 {
		file = fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	} else {
		file = fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.User)
	}

	robot.writeLog(file, fmt.Sprintf("%s: %s", ev.User, ev.Data))
}

func (robot *Bot) logEventJOIN(ev *Event) {
	file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	robot.writeLog(file, fmt.Sprintf("%s has joined %s", ev.User, ev.Channel))
}

func (robot *Bot) logEventPART(ev *Event) {
	file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	robot.writeLog(file, fmt.Sprintf("%s has left %s", ev.User, ev.Channel))
}

func (robot *Bot) logEventKICK(ev *Event) {
	file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	robot.writeLog(file, fmt.Sprintf("%s has been kicked from %s by %s", ev.Data, ev.Channel, ev.User))
}

// LogEvent logs events
func (robot *Bot) LogEvent(ev *Event) {
	if robot.Config.Logs.Enable {
		switch ev.Type {
		case E_PRIVMSG:
			robot.logEventPRIVMSG(ev)
		case E_JOIN:
			robot.logEventJOIN(ev)
		case E_PART:
			robot.logEventPART(ev)
		case E_KICK:
			robot.logEventKICK(ev)
		}
	}
}
