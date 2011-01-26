package gorobot

import (
	"os"
	"netchan"
	"fmt"
)

type GoRobot struct {
	Irc* Irc
	Exp *netchan.Exporter
	Modules map[string] chan Event
	Actions chan Action
}

// Creates a new robot from a configuration file
func NewGoRobot() *GoRobot {
	robot := GoRobot{
		Irc: NewIrc("m1ch3ld3uX"),
		Modules: make(map[string] chan Event),
		Exp: InitExport("localhost:12345"),
	}
	robot.Actions = ExportActions(robot.Exp)
	robot.Irc.Connect("irc.freenode.net:6667", "freenode")
	return &robot
}

func (robot *GoRobot) SendEvent(event *Event) {
	for _, chev := range robot.Modules {
		go func (chev chan Event, event Event) {
			chev <- event
		} (chev, *event);
	}
}

// Based on PING events from servers, ugly but enough for now
func (robot *GoRobot) Cron() {
	robot.Irc.CleanConversations()
}

func (robot *GoRobot) HandleEvent(s *Server, event *Event) {
	switch event.Type {
	case E_PING :
		s.SendMeRaw <- fmt.Sprintf("PONG :%s\r\n", event.Data)
		robot.Cron()
	case E_NOTICE :
		if !s.AuthSent {
			s.SendMeRaw <- fmt.Sprintf("NICK %s\r\n", s.BotName)
			s.SendMeRaw <- fmt.Sprintf("USER %s 0.0.0.0 0.0.0.0 :%s\r\n", s.BotName, s.BotName)
			s.AuthSent = true
		}
	case E_PRIVMSG :
		if s.Channels[event.Channel] != nil {
			event.AdminCmd = s.Channels[event.Channel].Master
		}
	}

	robot.SendEvent(event)
	if event.CmdId == 1 {
		switch event.Server {
		case "freenode":
			robot.Irc.JoinChannel("freenode", "##testpikaplop", true, "")
		}
		return
	}
}

func (robot *GoRobot) HandleError(e os.Error) {
}

func (robot *GoRobot) NewModule(ac *Action) {
	robot.Modules[ac.Data] = ExportEvents(robot.Exp, ac.Data)
}

func (robot *GoRobot) HandleAction(ac *Action) {
	if ac.Type == A_RAW {
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
	case A_NEWMODULE:
		robot.NewModule(ac)
	case A_SAY:
		robot.Irc.Say(ac)
	case A_JOIN:
		robot.Irc.Join(ac)
	case A_PART:
		robot.Irc.Part(ac)
	case A_KICK:
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
