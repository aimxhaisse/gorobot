package gorobot

import (
	"api"
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
	"strings"
)

// IRC Channel
type Channel struct {
	Config	ConfigChannel		// Channel configuration
	Say	map[int] chan string	// map of channels to talk, each channel has a priority
	Server	*Server			// server when the channel is
	Destroy	chan int		// force the destruction of the IRC channel
}

// IRC Conversation
type Conversation struct {
	LastUpdate	int64		// inactive conversations will be destroyed
	Say		map[int] chan string // map of channels to talk, each channel has a priority
	Destroy		chan int	// force the destruction of the conversation
}

// IRC Server
type Server struct {
	Config		ConfigServer
	Channels	map[string] *Channel // IRC channels where the bot is
	Conversations	map[string] Conversation // opened conversations
	SendMeRaw	chan string	// channel to send raw commands to the server
	AuthSent	bool		// has the authentication been sent?
	socket		net.Conn	// socket to the server
}

// IRC Bot
type Irc struct {
	Events      chan api.Event	// Events are written here
	Errors      chan os.Error	// Useless for now
	Servers	    map[string] *Server	// Servers where the bot is connected
}

// Instanciate a new IRC bot
func NewIrc() *Irc {
	b := Irc{
		Events: make(chan api.Event),
		Errors: make(chan os.Error),
		Servers: make(map[string] *Server),
	}
	return &b
}

// Returns nil or the server which alias is srv
func (irc *Irc) GetServer(srv string) (*Server) {
	return irc.Servers[srv]
}

// Returns nil or the irc channel from the server
func (irc *Irc) GetChannel(srv string, channel string) (*Channel) {
	var server = irc.GetServer(srv)
	if server != nil {
		return server.Channels[channel]
	}
	return nil
}

// Periodically called to remove inactive conversations
func (irc *Irc) CleanConversations() {
	current_time := time.Seconds()
	for _, s := range irc.Servers {
		for k, conv := range s.Conversations {
			if (current_time - conv.LastUpdate) > 600 {
				conv.Destroy <- 42
				<- conv.Destroy
				s.Conversations[k] = conv, false
			}
		}
	}
}

// Extract events from the server
func reader(srv_name string, socket net.Conn, chev chan api.Event) {
	r := bufio.NewReader(socket)
	for {
		var err os.Error
		var p []byte
		if p, err = r.ReadSlice('\n'); err != nil {
			return
		}

		line := strings.TrimRight(string(p), "\r\t\n")
		ev := ExtractEvent(line)
		if ev != nil {
			ev.Server = srv_name
			fmt.Printf("\x1b[1;36m%s\x1b[0m\n", line)
			chev <- *ev
		}
	}
}

// Write raw commands to the server
func writer(socket net.Conn, chsend chan string) {
	var str string
	var raw []byte

	for {
		str = <- chsend
		raw = []byte(str)
		fmt.Printf("\x1b[1;35m%s\x1b[0m\n", strings.TrimRight(str, "\r\t\n"))
		if _, err := socket.Write(raw); err != nil {
			return
		}
	}
}

// Connect to a new server
func (irc *Irc) Connect(c ConfigServer) bool {
	if irc.GetServer(c.Name) != nil {
		fmt.Printf("Already connected to that server [%s]\n", c.Host)
		return false
	}
	fmt.Printf("Connecting to [%s]\n", c.Host)
	conn, err := net.Dial("tcp", "", c.Host)
	if err != nil {
		irc.Errors <- err
		return false
	}
	srv := Server{
		Config: c,
		Channels: make(map[string] *Channel),
		SendMeRaw: make(chan string),
		Conversations: make(map[string] Conversation),
		socket: conn,
	}
	irc.Servers[c.Name] = &srv
	go reader(srv.Config.Name, srv.socket, irc.Events)
	go writer(srv.socket, srv.SendMeRaw)
	return true
}

// Say something to a channel, handling excess flood
func sayToChannel(after *int64, ahead *int64, before *int64, chout chan string, say string, target string) {
	*after = time.Nanoseconds()
	*ahead -= (*after - *before)
	if *ahead < 0 {
		*ahead = 0
	} else if *ahead > 10e9 {
		time.Sleep(*ahead - 10e9)
		*ahead = 10e9
	}
	go func (privmsg string) {
		chout <- privmsg
	} (fmt.Sprintf("PRIVMSG %s :%s\r\n", target, say))
	*ahead += 2e9
	*before = time.Nanoseconds()
}

// Wait for activity on say channels
func talkChannel(target string, chin *map[int] chan string, chout chan string, destroy chan int) {
	var after int64 = 0
	var ahead int64 = 0
	before := time.Nanoseconds()
	// "while the timer is less than ten seconds ahead of the current time, parse any
	// present messages and penalize the client by 2 seconds for each message" (doc irssi)
	for {
		select {
		case <- destroy:
			destroy <- 42
			return
		case say := <- (*chin)[api.PRIORITY_HIGH]:
			sayToChannel(&after, &ahead, &before, chout, say, target)
		case say := <- (*chin)[api.PRIORITY_MEDIUM]:
			sayToChannel(&after, &ahead, &before, chout, say, target)
		case say := <- (*chin)[api.PRIORITY_LOW]:
			sayToChannel(&after, &ahead, &before, chout, say, target)
		}
	}
}

// Join a new IRC channel
func (irc *Irc) JoinChannel(conf ConfigChannel, irc_server string) {
	s := irc.GetServer(irc_server)
	if s == nil {
		return
	} else if irc.GetChannel(irc_server, conf.Name) != nil {
		fmt.Printf("Channel %s already exists on %s\n", conf.Name, irc_server)
		return
	}

	if len(conf.Password) > 0 {
		s.SendMeRaw <- fmt.Sprintf("JOIN %s %s\r\n", conf.Name, conf.Password)
	} else {
		s.SendMeRaw <- fmt.Sprintf("JOIN %s\r\n", conf.Name)
	}

	c := Channel{
		Config: conf,
		Server: s,
		Destroy: make(chan int),
		Say: make(map[int] chan string),
	}
	c.Say[api.PRIORITY_LOW] = make(chan string)
	c.Say[api.PRIORITY_MEDIUM] = make(chan string)
	c.Say[api.PRIORITY_HIGH] = make(chan string)
	s.Channels[conf.Name] = &c
	fmt.Printf("Having joined %s on %s\n", conf.Name, irc_server)
	go talkChannel(conf.Name, &c.Say, s.SendMeRaw, c.Destroy)
}

// Join a channel with a default configuration
func (irc *Irc) Join(ac *api.Action) {

	conf := ConfigChannel{
		Master: false,
		Name: ac.Channel,
	}

	irc.JoinChannel(conf, ac.Server)
}

// Kick someone from an IRC channel
func (irc *Irc) Kick(ac *api.Action) {
	c := irc.GetChannel(ac.Server, ac.Channel)
	s := irc.GetServer(ac.Server)

	if c != nil {
		if len(ac.Data) > 0 {
			s.SendMeRaw <- fmt.Sprintf("KICK %s %s :%s\r\n", c.Config.Name, ac.User, ac.Data)
		} else {
			s.SendMeRaw <- fmt.Sprintf("KICK %s %s\r\n", c.Config.Name, ac.User)
		}
	}
}

// Leave an IRC channel
func (irc *Irc) Part(ac *api.Action) {
	s := irc.GetServer(ac.Server)
	c := irc.GetChannel(ac.Server, ac.Channel)

	if c != nil {
		if len(ac.Data) > 0 {
			s.SendMeRaw <- fmt.Sprintf("PART %s :%s\r\n", ac.Channel, ac.Data)
		} else {
			s.SendMeRaw <- fmt.Sprintf("PART %s\r\n", ac.Channel)
		}
		c.Destroy <- 42
		<- c.Destroy
		s.Channels[ac.Channel] = s.Channels[ac.Channel], false
		fmt.Printf("Having left channel %s on %s\n", ac.Channel, ac.Server)
	}
}

// Create a new conversation, with the same behavior as a channel
func (irc *Irc) CreateNewConversation(ac *api.Action, server *Server) (Conversation) {
	conv := Conversation{
		Say: make(map[int] chan string),
		Destroy: make(chan int),
	}

	conv.Say[api.PRIORITY_LOW] = make(chan string)
	conv.Say[api.PRIORITY_MEDIUM] = make(chan string)
	conv.Say[api.PRIORITY_HIGH] = make(chan string)
	server.Conversations[ac.User] = conv

	go talkChannel(ac.User, &conv.Say, server.SendMeRaw, conv.Destroy)

	return conv
}

// Say something to a channel or to a conversation, create a new conversation
// if it does not exists yet
func (irc *Irc) Say(ac *api.Action) {
	var server = irc.Servers[ac.Server]
	if server != nil {
		if len(ac.Channel) > 0 {
			channel := server.Channels[ac.Channel]
			if channel != nil {
				if len(ac.User) > 0 {
					ac.Data = fmt.Sprintf("%s: %s", ac.User, ac.Data)
				}
				go func(data string, p int) {
					channel.Say[p] <- data
				}(ac.Data, ac.Priority);
			}
		} else {
			var ok bool
			var conv Conversation

			conv, ok = server.Conversations[ac.User]
			if !ok {
				conv = irc.CreateNewConversation(ac, server)
			}
			conv.LastUpdate = time.Seconds()
			go func(data string, p int) {
				conv.Say[p] <- data
			}(ac.Data, ac.Priority);
		}
	}
}
