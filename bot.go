package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Bot struct {
	Config        *Config               // Main config of the bot
	LogMap        map[string]*os.File   // Log files
	LogLock       sync.Mutex            // Mutex for log
	Irc           *Irc                  // Current IRC state
	Modules       map[string]chan Event // Loaded modules
	Actions       chan Action           // Read actions from modules
	WebAPIActions chan Action           // These are special actions that should be handled by the WebAPI
}

// NewBot creates a new IRC bot with the given config
func NewBot(cfg *Config) *Bot {
	b := Bot{
		Config:        cfg,
		LogMap:        make(map[string]*os.File),
		LogLock:       sync.Mutex{},
		Irc:           NewIrc(),
		Modules:       make(map[string]chan Event),
		Actions:       make(chan Action),
		WebAPIActions: make(chan Action),
	}
	b.initLog(b.Config.Logs)
	b.Irc.Connect(b.Config.Servers)

	b.Modules["broadcast"] = make(chan Event)
	go Broadcast(b.Actions, b.Modules["broadcast"], cfg.Broadcast)

	b.Modules["scripts"] = make(chan Event)
	go Scripts(b.Actions, b.Modules["scripts"], &b, cfg.Scripts)

	b.Modules["markov"] = make(chan Event)
	go Markov(b.Actions, b.Modules["markov"], cfg.Markov)

	go WebAPI(&cfg.WebAPI, b.Irc.Events, b.WebAPIActions)

	return &b
}

// Run is the main loop of the IRC bot
func (b *Bot) Run() {
	for {
		select {
		case _ = <-time.Tick(60 * time.Second):
			b.Irc.AutoReconnect()
		case action, ok := <-b.Actions:
			if !ok {
				log.Printf("action channel closed, bye bye")
				return
			}
			b.handleAction(&action)
		case event, ok := <-b.Irc.Events:
			if !ok {
				log.Printf("event channel closed, bye bye")
			}
			srv := b.Irc.GetServer(event.Server)
			if srv != nil {
				b.handleEvent(srv, &event)
			} else if event.Server == b.Config.WebAPI.HTTPServerName {
				log.Printf("")
				// Dispatch the event to modules
				for _, chev := range b.Modules {
					go func(chev chan Event, event Event) {
						chev <- event
					}(chev, event)
				}
			}
		}
	}
}

// initLog initializes the log directory
func (b *Bot) initLog(config ConfigLogs) {
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

// autoJoin joins the configured chans for the given server
func (b *Bot) autoJoin(s string) {
	serv := b.Irc.GetServer(s)
	if serv != nil {
		for k, _ := range serv.Config.Channels {
			serv.JoinChannel(k)
		}
	}
}

// HandleEvent processes an event from a server
func (b *Bot) handleEvent(serv *Server, event *Event) {
	switch event.Type {
	case E_KICK:
		if serv.Config.Nickname == event.Data && b.Config.AutoRejoinOnKick {
			serv.JoinChannel(event.Channel)
		}
	case E_PING:
		serv.SendMeRaw[PRIORITY_HIGH] <- fmt.Sprintf("PONG :%s\r\n", event.Data)
	case E_NOTICE:
		if event.CmdId == 1 {
			b.autoJoin(serv.Config.Name)
		}
	case E_DISCONNECT:
		serv.Disconnect()
	case E_PRIVMSG:
		if _, ok := serv.Config.Channels[event.Channel]; ok == true {
			event.AdminCmd = serv.Config.Channels[event.Channel].Master
		}
	}
	// Dispatch the event to modules
	for _, chev := range b.Modules {
		go func(chev chan Event, event Event) {
			chev <- event
		}(chev, *event)
	}
	b.LogEvent(event)
}

// NewModule registers a new module
func (b *Bot) newModule(ac *Action) {
	b.Modules[ac.Data] = make(chan Event)
}

// HandleAction processes an action from a module
func (b *Bot) handleAction(ac *Action) {
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

	if ac.Server == b.Config.WebAPI.HTTPServerName {
		b.WebAPIActions <- *ac
	} else {
		switch ac.Type {
		case A_SAY:
			if serv := b.Irc.GetServer(ac.Server); serv != nil {
				serv.Say(ac)
			}
		case A_JOIN:
			if serv := b.Irc.GetServer(ac.Server); serv != nil {
				serv.JoinChannel(ac.Channel)
			}
		case A_PART:
			if serv := b.Irc.GetServer(ac.Server); serv != nil {
				serv.LeaveChannel(ac.Channel, ac.Data)
			}
		case A_KICK:
			if serv := b.Irc.GetServer(ac.Server); serv != nil {
				serv.KickUser(ac.Channel, ac.User, ac.Data)
			}
		case A_NAMES:
			if serv := b.Irc.GetServer(ac.Server); serv != nil {
				serv.Names(ac)
			}
		}
	}

	b.LogAction(ac)
}
