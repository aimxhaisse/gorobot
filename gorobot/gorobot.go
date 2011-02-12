package gorobot

import (
	"botapi"
	"os"
	"netchan"
	"fmt"
	"log"
)

type GoRobot struct {
	Config *Config
	LogMap map[string] *os.File
	Irc* Irc
	Exp *netchan.Exporter
	Modules map[string] chan botapi.Event
	Actions chan botapi.Action
}

// Creates a new robot from a configuration file, automatically
// connect to servers listed in the configuration file
func NewGoRobot(config string) *GoRobot {
	robot := GoRobot{
		Config: NewConfig(config),
		LogMap: make(map[string] *os.File),
		Irc: NewIrc(),
		Modules: make(map[string] chan botapi.Event),
	}
	robot.Exp = botapi.InitExport(robot.Config.Module.Interface)
	robot.Actions = botapi.ExportActions(robot.Exp)
	for  k, v := range robot.Config.Servers {
		robot.Irc.Connect(k, v)
	}
	return &robot
}

func (robot *GoRobot) SendEvent(event *botapi.Event) {
	for _, chev := range robot.Modules {
		chev <- *event
	}
	robot.LogEvent(event)
}

// Based on PING events from servers, ugly but enough for now
func (robot *GoRobot) Cron() {
	robot.Irc.CleanConversations()
	robot.LogStatistics()
}

// Autojoin channels on a given server
func (robot *GoRobot) autoJoin(s string) {
	srv := robot.Irc.GetServer(s)
	if srv != nil {
		for k, v := range srv.Config.Channels {
			robot.Irc.JoinChannel(v, s, k)
		}
	}
}

// Handle a notice
func (robot *GoRobot) HandleNotice(s *Server, event *botapi.Event) {
	switch event.CmdId {
	case 1:
		robot.autoJoin(s.Config.Name)
	case 353:
		robot.Irc.AddUsersToChannel(s, event)
	}
}

// Handle an event from a server
func (robot *GoRobot) HandleEvent(s *Server, event *botapi.Event) {
	switch event.Type {
	case botapi.E_KICK :
		if s.Config.Nickname == event.Data {
			robot.Irc.DestroyChannel(event.Server, event.Channel)
		} else {
			robot.Irc.UserLeft(event)
		}
	case botapi.E_PING :
		s.SendMeRaw <- fmt.Sprintf("PONG :%s\r\n", event.Data)
		robot.Cron()
	case botapi.E_NOTICE :
		robot.HandleNotice(s, event)
	case botapi.E_PRIVMSG :
		if s.Channels[event.Channel] != nil {
			event.AdminCmd = s.Channels[event.Channel].Config.Master
		}
	case botapi.E_JOIN :
		robot.Irc.UserJoined(event)
	case botapi.E_PART :
		robot.Irc.UserLeft(event)
	case botapi.E_QUIT :
		robot.Irc.UserQuit(event)
	}
	robot.SendEvent(event)
}

func (robot *GoRobot) HandleError(e os.Error) {
}

func (robot *GoRobot) NewModule(ac *botapi.Action) {
	robot.Modules[ac.Data] = botapi.ExportEvents(robot.Exp, ac.Data)
}

func (robot *GoRobot) HandleAction(ac *botapi.Action) {
	// if the command is RAW, we need to parse it first to be able
	// to correctly handle it.
	if ac.Type == botapi.A_RAW {
		new_action := ExtractAction(ac)
		if new_action != nil {
			p := ac.Priority
			*ac = *new_action
			ac.Priority = p
		} else {
			log.Printf("raw command ignored [%s]\n", ac.Raw)
			return
		}
	}

	switch ac.Type {
	case botapi.A_NEWMODULE:
		robot.NewModule(ac)
	case botapi.A_SAY:
		robot.Irc.Say(ac)
	case botapi.A_JOIN:
		robot.Irc.Join(ac)
	case botapi.A_PART:
		robot.Irc.Part(ac)
	case botapi.A_KICK:
		robot.Irc.Kick(ac)
	}
}

func (robot *GoRobot) Run() {
	for {
		select {
		case action := <-robot.Actions:
			robot.HandleAction(&action)
		case event := <-robot.Irc.Events:
			srv := robot.Irc.GetServer(event.Server)
			if srv != nil {
				robot.HandleEvent(srv, &event)
			}
		case err := <-robot.Irc.Errors:
			robot.HandleError(err)
		}
	}
}
