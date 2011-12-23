package main

import (
	"api"
	"fmt"
	"log"
	"old/netchan"
	"os"
	"os/exec"
	"time"
)

type GoRobot struct {
	Config  *Config
	LogMap  map[string]*os.File
	Irc     *Irc
	Exp     *netchan.Exporter
	Modules map[string]chan api.Event
	Actions chan api.Action
}

func NewGoRobot(config string) *GoRobot {
	robot := GoRobot{
		Config:  NewConfig(config),
		LogMap:  make(map[string]*os.File),
		Irc:     NewIrc(),
		Modules: make(map[string]chan api.Event),
	}
	robot.Exp = api.InitExport(robot.Config.Module.Interface)
	robot.Actions = api.ExportActions(robot.Exp)
	robot.InitLog(robot.Config.Logs)
	robot.Irc.Connect(robot.Config.Servers)
	return &robot
}

func (robot *GoRobot) InitLog(config ConfigLogs) {
	if config.Enable == true {
		_, err := os.Open(config.Directory)
		if err != nil {
			err = os.Mkdir(config.Directory, 0755)
			if err != nil {
				log.Fatalf("Can't create log directory: %v", err)
			}
		}
	}
}

func (robot *GoRobot) SendEvent(event *api.Event) {
	for _, chev := range robot.Modules {
		go func(chev chan api.Event, event api.Event) {
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

func (robot *GoRobot) HandleNotice(serv *Server, event *api.Event) {
	switch event.CmdId {
	case 1:
		robot.AutoJoin(serv.Config.Name)
	}
}

func (robot *GoRobot) HandleEvent(serv *Server, event *api.Event) {
	switch event.Type {
	case api.E_KICK:
		if serv.Config.Nickname == event.Data && robot.Config.AutoRejoinOnKick {
			serv.JoinChannel(event.Channel)
		}
	case api.E_PING:
		serv.SendMeRaw[api.PRIORITY_HIGH] <- fmt.Sprintf("PONG :%s\r\n", event.Data)
	case api.E_NOTICE:
		robot.HandleNotice(serv, event)
	case api.E_DISCONNECT:
		serv.Disconnect()
	case api.E_PRIVMSG:
		if _, ok := serv.Config.Channels[event.Channel]; ok == true {
			event.AdminCmd = serv.Config.Channels[event.Channel].Master
		}
	}
	robot.SendEvent(event)
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
			log.Printf("raw command ignored [%s]", ac.Raw)
			return
		}
	}

	switch ac.Type {
	case api.A_NEWMODULE:
		robot.NewModule(ac)
	case api.A_SAY:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.Say(ac)
		}
	case api.A_JOIN:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.JoinChannel(ac.Channel)
		}
	case api.A_PART:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.LeaveChannel(ac.Channel, ac.Data)
		}
	case api.A_KICK:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.KickUser(ac.Channel, ac.User, ac.Data)
		}
	}
}

func ScheduleCron(cron chan int, timeout int64) {
	if timeout <= 0 {
		timeout = 60
	}
	duration := time.Duration(timeout)
	for {
		cron <- 42
		time.Sleep(duration * time.Second)
	}
}

func (robot *GoRobot) AutoRunModules() {
	if robot.Config.Module.AutoRunModules {
		for _, module := range robot.Config.Module.AutoRun {
			log.Printf("launching %s", module)
			go func(module string) {
				cmd := exec.Command(module)
				err := cmd.Run()
				if err == nil {
					cmd.Wait()
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
