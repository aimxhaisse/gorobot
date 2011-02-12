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

// Returns nil or the server which alias is srv
func (irc *Irc) GetServer(srv string) (*Server) {
	return irc.Servers[srv]
}

// Connect to a new server
func (irc *Irc) Connect(conf *ConfigServer) bool {
	log.Printf("connecting to [%s]\n", conf.Host)
	if irc.GetServer(conf.Name) != nil {
		log.Printf("already connected to that server [%s]", conf.Host)
		return false
	}
	irc.Servers[conf.Name] = NewServer(conf, irc.Events)
	return true
}
