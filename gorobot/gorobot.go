package gorobot

import (
	"api"
	"os"
	"netchan"
	"fmt"
)

type GoRobot struct {
	Config *Config
	Irc* Irc
	Exp *netchan.Exporter
	Modules map[string] chan api.Event
	Actions chan api.Action
}

// Creates a new robot from a configuration file, automatically
// connect to servers listed in the configuration file
func NewGoRobot(config string) *GoRobot {
	robot := GoRobot{
		Config: NewConfig(config),
		Irc: NewIrc(),
		Modules: make(map[string] chan api.Event),
	}
	robot.Exp = api.InitExport(robot.Config.Module.Interface)
	robot.Actions = api.ExportActions(robot.Exp)
	for  _, v := range robot.Config.Servers {
		robot.Irc.Connect(v)
	}
	return &robot
}

func (robot *GoRobot) SendEvent(event *api.Event) {
	for _, chev := range robot.Modules {
		go func (chev chan api.Event, event api.Event) {
			chev <- event
		} (chev, *event);
	}
}

// Based on PING events from servers, ugly but enough for now
func (robot *GoRobot) Cron() {
	robot.Irc.CleanConversations()
}

// Autojoin channels on a given server
func (robot *GoRobot) autoJoin(s string) {
	srv := robot.Irc.GetServer(s)
	if srv != nil {
		for _, v := range srv.Config.Channels {
			robot.Irc.JoinChannel(v, s)
		}
	}
}

func (robot *GoRobot) HandleEvent(s *Server, event *api.Event) {
	switch event.Type {
	case api.E_PING :
		s.SendMeRaw <- fmt.Sprintf("PONG :%s\r\n", event.Data)
		robot.Cron()
	case api.E_NOTICE :
		if !s.AuthSent {
			s.SendMeRaw <- fmt.Sprintf("NICK %s\r\n", s.Config.Nickname)
			s.SendMeRaw <- fmt.Sprintf("USER %s 0.0.0.0 0.0.0.0 :%s\r\n",
				s.Config.Username, s.Config.Realname)
			s.AuthSent = true
		}
	case api.E_PRIVMSG :
		if s.Channels[event.Channel] != nil {
			event.AdminCmd = s.Channels[event.Channel].Config.Master
		}
	}
	if event.CmdId == 1 {
		robot.autoJoin(event.Server)
	}
	robot.SendEvent(event)
}

func (robot *GoRobot) HandleError(e os.Error) {
}

func (robot *GoRobot) NewModule(ac *api.Action) {
	robot.Modules[ac.Data] = api.ExportEvents(robot.Exp, ac.Data)
}

func (robot *GoRobot) HandleAction(ac *api.Action) {
	// if the command is RAW, we need to parse it first to be able
	// to correctly handle it.
	if ac.Type == api.A_RAW {
		new_action := ExtractAction(ac)
		if new_action != nil {
			p := ac.Priority
			*ac = *new_action
			ac.Priority = p
		} else {
			fmt.Printf("Raw command ignored [%s]\n", ac.Raw)
			return
		}
	}

	switch ac.Type {
	case api.A_NEWMODULE:
		robot.NewModule(ac)
	case api.A_SAY:
		robot.Irc.Say(ac)
	case api.A_JOIN:
		robot.Irc.Join(ac)
	case api.A_PART:
		robot.Irc.Part(ac)
	case api.A_KICK:
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
