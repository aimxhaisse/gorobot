package gorobot

import (
	"botapi"
	"os"
	"netchan"
	"fmt"
	"log"
	"time"
	"exec"
)

type GoRobot struct {
	Config  *Config
	LogMap  map[string]*os.File
	Irc     *Irc
	Exp     *netchan.Exporter
	Modules map[string]chan botapi.Event
	Actions chan botapi.Action
}

func NewGoRobot(config string) *GoRobot {
	robot := GoRobot{
		Config:  NewConfig(config),
		LogMap:  make(map[string]*os.File),
		Irc:     NewIrc(),
		Modules: make(map[string]chan botapi.Event),
	}
	robot.Exp = botapi.InitExport(robot.Config.Module.Interface)
	robot.Actions = botapi.ExportActions(robot.Exp)
	robot.Irc.Connect(robot.Config.Servers)
	return &robot
}

func (robot *GoRobot) SendEvent(event *botapi.Event) {
	for _, chev := range robot.Modules {
		go func(chev chan botapi.Event, event botapi.Event) {
			chev <- event
		}(chev, *event)
	}
	robot.LogEvent(event)
}

func (robot *GoRobot) Cron() {
	robot.LogStatistics()
	robot.Irc.AutoReconnect()
}

func (robot *GoRobot) AutoJoin(s string) {
	serv := robot.Irc.GetServer(s)
	if serv != nil {
		for k, _ := range serv.Config.Channels {
			serv.JoinChannel(k)
		}
	}
}

func (robot *GoRobot) HandleNotice(serv *Server, event *botapi.Event) {
	switch event.CmdId {
	case 1:
		robot.AutoJoin(serv.Config.Name)
	}
}

func (robot *GoRobot) HandleEvent(serv *Server, event *botapi.Event) {
	switch event.Type {
	case botapi.E_KICK:
		if serv.Config.Nickname == event.Data && robot.Config.AutoRejoinOnKick {
			serv.JoinChannel(event.Channel)
		}
	case botapi.E_PING:
		serv.SendMeRaw[botapi.PRIORITY_HIGH] <- fmt.Sprintf("PONG :%s\r\n", event.Data)
	case botapi.E_NOTICE:
		robot.HandleNotice(serv, event)
	case botapi.E_DISCONNECT:
		serv.Disconnect()
	case botapi.E_PRIVMSG:
		if _, ok := serv.Config.Channels[event.Channel]; ok == true {
			event.AdminCmd = serv.Config.Channels[event.Channel].Master
		}
	}
	robot.SendEvent(event)
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
			log.Printf("raw command ignored [%s]", ac.Raw)
			return
		}
	}

	switch ac.Type {
	case botapi.A_NEWMODULE:
		robot.NewModule(ac)
	case botapi.A_SAY:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.Say(ac)
		}
	case botapi.A_JOIN:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.JoinChannel(ac.Channel)
		}
	case botapi.A_PART:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.LeaveChannel(ac.Channel, ac.Data)
		}
	case botapi.A_KICK:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.KickUser(ac.Channel, ac.User, ac.Data)
		}
	}
}

func ScheduleCron(cron chan int, timeout int64) {
	if timeout <= 0 {
		timeout = 60
	}
	timeout = timeout * 1e9
	for {
		cron <- 42
		time.Sleep(timeout)
	}
}

func (robot *GoRobot) AutoRunModules() {
	if robot.Config.Module.AutoRunModules {
		for _, module := range robot.Config.Module.AutoRun {
			log.Printf("launching %s", module)
			go func(module string) {
				cmd, err := exec.Run(module, []string{module}, []string{}, "",
					exec.DevNull, exec.PassThrough, exec.PassThrough)
				if err == nil {
					cmd.Wait(0)
				} else {
					log.Printf("can't run module %s: %v", module, err)
				}
			}(module)
		}
	}
}

func (robot *GoRobot) Run() {
	cron := make(chan int)

	robot.AutoRunModules()
	go ScheduleCron(cron, robot.Config.CronTimeout)

	for {
		select {
		case _ = <-cron:
			robot.Cron()
		case action, ok := <-robot.Actions:
			if !ok {
				log.Printf("action channel closed, bye bye")
				return
			}
			robot.HandleAction(&action)
		case event, ok := <-robot.Irc.Events:
			if !ok {
				log.Printf("event channel closed, bye bye")
			}
			srv := robot.Irc.GetServer(event.Server)
			if srv != nil {
				robot.HandleEvent(srv, &event)
			}
		}
	}
}
