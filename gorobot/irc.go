package gorobot

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
	"strings"
)

// IRC Channel
type Channel struct {
	Master	bool // master IRC channel can issue admin commands
	Name	string // name of the channel
	Say	map[int] chan string // map of channels to talk, each channel has a priority
	Server	*Server // server when the channel is
	Destroy	chan int // force the destruction of the IRC channel
}

// IRC Conversation
type Conversation struct {
	LastUpdate	int64 // inactive conversations will be destroyed
	Say		map[int] chan string // map of channels to talk, each channel has a priority
	Destroy		chan int // force the destruction of the conversation
}

// IRC Server
type Server struct {
	BotName		string // name of the bot on the server
	Channels	map[string] *Channel // IRC channels where the bot is
	Conversations	map[string] Conversation // opened conversations
	AuthSent	bool // has the authentication been sent?
	Host		string // hostname of the server
	Name		string // alias to hostname
	SendMeRaw	chan string // channel to send raw commands to the server
	socket		net.Conn // socket to the server
}

// IRC Bot
type Irc struct {
	DefaultName string
	Events      chan Event
	Errors      chan os.Error
	Servers	    map[string] *Server
}

// Instanciate a new IRC bot
func NewIrc(defname string) *Irc {
	b := Irc{
		DefaultName: defname,
		Events: make(chan Event),
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
func reader(srv_name string, socket net.Conn, chev chan Event) {
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
func (irc *Irc) Connect(srv_host string, srv_name string) bool {
	if irc.GetServer(srv_name) != nil {
		return false
	}
	fmt.Printf("Connecting to [%s]\n", srv_host)
	conn, err := net.Dial("tcp", "", srv_host)
	if err != nil {
		irc.Errors <- err
		return false
	}

	srv := Server{
		BotName: irc.DefaultName,
		Channels: make(map[string] *Channel),
		Host: srv_host,
		Name: srv_name,
		SendMeRaw: make(chan string),
		socket: conn,
		Conversations: make(map[string] Conversation),
	}
	irc.Servers[srv_name] = &srv

	go reader(srv.Name, srv.socket, irc.Events)
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
		case say := <- (*chin)[PRIORITY_HIGH]:
			sayToChannel(&after, &ahead, &before, chout, say, target)
		case say := <- (*chin)[PRIORITY_MEDIUM]:
			sayToChannel(&after, &ahead, &before, chout, say, target)
		case say := <- (*chin)[PRIORITY_LOW]:
			sayToChannel(&after, &ahead, &before, chout, say, target)
		}
	}
}

// Join a new IRC channel
func (irc *Irc) JoinChannel(irc_server string, irc_chan string, master bool, password string) {
	s := irc.GetServer(irc_server)

	if s == nil {
		return
	}
	if irc.GetChannel(irc_server, irc_chan) != nil {
		fmt.Printf("Channel %s already exists on %s\n", irc_chan, irc_server)
		return
	}

	if len(password) > 0 {
		s.SendMeRaw <- fmt.Sprintf("JOIN %s %s\r\n", irc_chan, password)
	} else {
		s.SendMeRaw <- fmt.Sprintf("JOIN %s\r\n", irc_chan)
	}

	c := Channel{
		Master: master,
		Name: irc_chan,
		Say: make(map[int] chan string),
		Server: s,
		Destroy: make(chan int),
	}
	c.Say[PRIORITY_LOW] = make(chan string)
	c.Say[PRIORITY_MEDIUM] = make(chan string)
	c.Say[PRIORITY_HIGH] = make(chan string)
	s.Channels[irc_chan] = &c
	fmt.Printf("Having joined %s on %s\n", irc_chan, irc_server)
	go talkChannel(irc_chan, &c.Say, s.SendMeRaw, c.Destroy)
}

func (irc *Irc) Join(ac *Action) {
	irc.JoinChannel(ac.Server, ac.Channel, false, "")
}

func (irc *Irc) Kick(ac *Action) {
	c := irc.GetChannel(ac.Server, ac.Channel)
	s := irc.GetServer(ac.Server)

	if c != nil {
		if len(ac.Data) > 0 {
			s.SendMeRaw <- fmt.Sprintf("KICK %s %s :%s\r\n", c.Name, ac.User, ac.Data)
		} else {
			s.SendMeRaw <- fmt.Sprintf("KICK %s %s\r\n", c.Name, ac.User)
		}
	}
}

func (irc *Irc) Part(ac *Action) {
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

func (irc *Irc) CreateNewConversation(ac *Action, server *Server) (Conversation) {
	conv := Conversation{
		Say: make(map[int] chan string),
		Destroy: make(chan int),
	}

	conv.Say[PRIORITY_LOW] = make(chan string)
	conv.Say[PRIORITY_MEDIUM] = make(chan string)
	conv.Say[PRIORITY_HIGH] = make(chan string)
	server.Conversations[ac.User] = conv

	go talkChannel(ac.User, &conv.Say, server.SendMeRaw, conv.Destroy)

	return conv
}

func (irc *Irc) Say(ac *Action) {
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
