package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Bot struct {
	Config  *Config               // Main config of the bot
	LogMap  map[string]*os.File   // Log files
	Irc     *Irc                  // Current IRC state
	Modules map[string]chan Event // Loaded modules
	Actions chan Action           // Read actions from modules
}

func NewBot(cfg *Config) *Bot {
	b := Bot{
		Config:  cfg,
		LogMap:  make(map[string]*os.File),
		Irc:     NewIrc(),
		Modules: make(map[string]chan Event),
		Actions: make(chan Action),
	}
	b.InitLog(b.Config.Logs)
	b.Irc.Connect(b.Config.Servers)
	return &b
}

func (robot *Bot) InitLog(config ConfigLogs) {
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

func (robot *Bot) SendEvent(event *Event) {
	for _, chev := range robot.Modules {
		go func(chev chan Event, event Event) {
			chev <- event
		}(chev, *event)
	}
	robot.LogEvent(event)
}

func (robot *Bot) Cron() {
	robot.Irc.AutoReconnect()
}

func (robot *Bot) AutoJoin(s string) {
	serv := robot.Irc.GetServer(s)
	if serv != nil {
		for k, _ := range serv.Config.Channels {
			serv.JoinChannel(k)
		}
	}
}

func (robot *Bot) HandleNotice(serv *Server, event *Event) {
	switch event.CmdId {
	case 1:
		robot.AutoJoin(serv.Config.Name)
	}
}

func (robot *Bot) HandleEvent(serv *Server, event *Event) {
	switch event.Type {
	case E_KICK:
		if serv.Config.Nickname == event.Data && robot.Config.AutoRejoinOnKick {
			serv.JoinChannel(event.Channel)
		}
	case E_PING:
		serv.SendMeRaw[PRIORITY_HIGH] <- fmt.Sprintf("PONG :%s\r\n", event.Data)
	case E_NOTICE:
		robot.HandleNotice(serv, event)
	case E_DISCONNECT:
		serv.Disconnect()
	case E_PRIVMSG:
		if _, ok := serv.Config.Channels[event.Channel]; ok == true {
			event.AdminCmd = serv.Config.Channels[event.Channel].Master
		}
	}
	robot.SendEvent(event)
}

func (robot *Bot) NewModule(ac *Action) {
	robot.Modules[ac.Data] = make(chan Event)
}

func (robot *Bot) HandleAction(ac *Action) {
	// if the command is RAW, we need to parse it first to be able
	// to correctly handle it.
	if ac.Type == A_RAW {
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
	case A_NEWMODULE:
		robot.NewModule(ac)
	case A_SAY:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.Say(ac)
		}
	case A_JOIN:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.JoinChannel(ac.Channel)
		}
	case A_PART:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.LeaveChannel(ac.Channel, ac.Data)
		}
	case A_KICK:
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

func (robot *Bot) Run() {
	cron := make(chan int)

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
