package gorobot

import (
	"botapi"
	"log"
	"os"
)

// IRC Bot
type Irc struct {
	Events      chan botapi.Event	// Events are written here
	Errors      chan os.Error	// Useless for now
	Servers	    map[string] *Server	// Servers where the bot is connected
}

// Instanciate a new IRC bot
func NewIrc() *Irc {
	b := Irc{
		Events: make(chan botapi.Event),
		Errors: make(chan os.Error),
		Servers: make(map[string] *Server),
	}
	return &b
}

// Returns nil or the server which alias is serv
func (irc *Irc) GetServer(serv string) (*Server) {
	result, ok := irc.Servers[serv]
	if ok == true && result.Connected == true {
		return result
	}
	return nil
}

// Connect to a new server
func (irc *Irc) Connect(servers map[string] *ConfigServer) {
	for  k, conf := range servers {
		conf.Name = k
		if irc.GetServer(conf.Name) != nil {
			log.Printf("already connected to that server [%s]", conf.Host)
			return
		}
		irc.Servers[conf.Name] = NewServer(conf, irc.Events)
	}
}

func (irc *Irc) AutoReconnect() {
	for _, serv := range irc.Servers {
		if serv.Connected == false {
			serv.TryReconnect(irc.Events)
		}
	}
}