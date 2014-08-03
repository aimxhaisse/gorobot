package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Writes the string to the log file, creates the log file if it doesn't exists
func (robot *Bot) writeLog(file string, what string, msg string) {
	currentTime := time.Now()
	strTime := currentTime.String()
	robot.LogLock.Lock()
	defer robot.LogLock.Unlock()
	fh, ok := robot.LogMap[file]
	if !ok {
		fh, _ = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if fh == nil {
			log.Printf("Warning: can't create file %s\n", file)
			return
		}
		robot.LogMap[file] = fh
	}
	fh.WriteString(fmt.Sprintf("%s %s %s\n", strTime, what, msg))
}

func (robot *Bot) logEventPRIVMSG(ev *Event) {
	var file string

	if len(ev.Channel) > 0 {
		file = fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	} else {
		file = fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.User)
	}

	robot.writeLog(file, "PRIVMSG", fmt.Sprintf("%s %s", ev.User, ev.Data))
}

func (robot *Bot) logEventJOIN(ev *Event) {
	file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	robot.writeLog(file, "JOIN", fmt.Sprintf("%s %s", ev.User, ev.Channel))
}

func (robot *Bot) logEventPART(ev *Event) {
	file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	robot.writeLog(file, "PART", fmt.Sprintf("%s %s", ev.User, ev.Channel))
}

func (robot *Bot) logEventKICK(ev *Event) {
	file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ev.Server, ev.Channel)
	robot.writeLog(file, "KICK", fmt.Sprintf("%s %s %s", ev.User, ev.Channel, ev.Data))
}

func (robot *Bot) LogCommand(server string, channel string, from string, cmd string) {
	if robot.Config.Logs.Enable {
		file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, server, channel)
		robot.writeLog(file, "CMD", fmt.Sprintf("%s %s", from, cmd))
	}
}

func (robot *Bot) logActionSAY(ac *Action) {
	if srv_cfg, ok := robot.Config.Servers[ac.Server]; ok {
		file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ac.Server, ac.Channel)
		robot.writeLog(file, "PRIVMSG", fmt.Sprintf("%s %s %s", srv_cfg.Nickname, ac.Channel, ac.Data))
	}
}

func (robot *Bot) logActionKICK(ac *Action) {
	if srv_cfg, ok := robot.Config.Servers[ac.Server]; ok {
		file := fmt.Sprintf("%s/%s-%s.log", robot.Config.Logs.Directory, ac.Server, ac.Channel)
		robot.writeLog(file, "KICK", fmt.Sprintf("%s %s %s", srv_cfg.Nickname, ac.Channel, ac.Data))
	}
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

// LogAction logs action
func (robot *Bot) LogAction(ac *Action) {
	if robot.Config.Logs.Enable {
		switch ac.Type {
		case A_SAY:
			robot.logActionSAY(ac)
		case A_KICK:
			robot.logActionKICK(ac)
		}
	}
}
