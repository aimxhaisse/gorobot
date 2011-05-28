package gorobot

import (
	"github.com/aimxhaisse/gorobot/api"
	"log"
	"os"
)

// IRC Bot
type Irc struct {
	Events  chan api.Event     // Events are written here
	Errors  chan os.Error      // Useless for now
	Servers map[string]*Server // Servers where the bot is connected
}

// NewIRC creates a  new IRC bot
func NewIrc() *Irc {
	b := Irc{
		Events:  make(chan api.Event),
		Errors:  make(chan os.Error),
		Servers: make(map[string]*Server),
	}
	return &b
}

// GetServer returns nil or the server whose alias is serv
func (irc *Irc) GetServer(serv string) *Server {
	result, ok := irc.Servers[serv]
	if ok == true && result.Connected == true {
		return result
	}
	return nil
}

// Connect creates a new IRC server and connects to it
func (irc *Irc) Connect(servers map[string]*ConfigServer) {
	for k, conf := range servers {
		conf.Name = k
		if irc.GetServer(conf.Name) != nil {
			log.Printf("already connected to that server [%s]", conf.Host)
			return
		}
		irc.Servers[conf.Name] = NewServer(conf, irc.Events)
	}
}

// AutoReconnect attempts to reconnect to server from which the IRC robot has been disconnected
func (irc *Irc) AutoReconnect() {
	for _, serv := range irc.Servers {
		if serv.Connected == false {
			serv.TryReconnect(irc.Events)
		}
	}
}
