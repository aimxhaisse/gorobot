package gorobot

import (
	"api"
	"bufio"
	"fmt"
	"net"
	"regexp"
	"os"
	"time"
	"strings"
)

// IRC Channel
type Channel struct {
	Config	ConfigChannel		// Channel configuration
	Say	map[int] chan string	// map of channels to talk, each channel has a priority
	Server	*Server			// server where the channel is
	Destroy	chan int		// force the destruction of the IRC channel
	Users	map[string] string	// Map of users with their mode
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
func (irc *Irc) Connect(alias string, c ConfigServer) bool {
	if irc.GetServer(alias) != nil {
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
	srv.Config.Name = alias
	irc.Servers[alias] = &srv
	go reader(srv.Config.Name, srv.socket, irc.Events)
	go writer(srv.socket, srv.SendMeRaw)
	return true
}

// Say something to a channel, handling excess flood
func sayToChannel(after *int64, ahead *int64, before *int64, chout chan string, say string, target string) {
	// "while the timer is less than ten seconds ahead of the current time, parse any
	// present messages and penalize the client by 2 seconds for each message" (doc irssi)
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
func (irc *Irc) JoinChannel(conf ConfigChannel, irc_server string, irc_chan string) {
	s := irc.GetServer(irc_server)
	if s == nil {
		return
	} else if irc.GetChannel(irc_server, irc_chan) != nil {
		fmt.Printf("Channel %s already exists on %s\n", irc_chan, irc_server)
		return
	}

	if len(conf.Password) > 0 {
		s.SendMeRaw <- fmt.Sprintf("JOIN %s %s\r\n", irc_chan, conf.Password)
	} else {
		s.SendMeRaw <- fmt.Sprintf("JOIN %s\r\n", irc_chan)
	}

	c := Channel{
		Config: conf,
		Server: s,
		Destroy: make(chan int),
		Users: make(map[string] string),
		Say: make(map[int] chan string),
	}
	c.Config.Name = irc_chan
	c.Say[api.PRIORITY_LOW] = make(chan string)
	c.Say[api.PRIORITY_MEDIUM] = make(chan string)
	c.Say[api.PRIORITY_HIGH] = make(chan string)
	s.Channels[irc_chan] = &c
	fmt.Printf("Having joined %s on %s\n", conf.Name, irc_server)
	go talkChannel(c.Config.Name, &c.Say, s.SendMeRaw, c.Destroy)
}

// A user has joined the channel
func (irc *Irc) UserJoined(ev *api.Event) {
	ch := irc.GetChannel(ev.Server, ev.Channel)
	if ch != nil {
		if _, ok := ch.Users[ev.User]; ok == false {
			ch.Users[ev.User] = "";
		}
	}
}


// A user has left the channel
func (irc *Irc) UserLeft(ev *api.Event) {
	ch := irc.GetChannel(ev.Server, ev.Channel)
	if ch != nil {
		ch.Users[ev.User] = ch.Users[ev.User], false
	}
}

// A user has left a server, lets remove it from each channel
func (irc *Irc) UserQuit(ev *api.Event) {
	s := irc.GetServer(ev.Server)
	if s != nil {
		for _, ch := range s.Channels {
			if _, ok := ch.Users[ev.User]; ok == true {
				ch.Users[ev.User] = ch.Users[ev.User], false;
			}
		}
	}
}

var re_event_userlist = regexp.MustCompile("^:[^ ]+ 353 [^:]+ . ([^ ]+) :(.*)")

// Add a list of users to a channel
func (irc *Irc) AddUsersToChannel(srv *Server, ev *api.Event) {
	m := re_event_userlist.FindStringSubmatch(ev.Raw)
	if len(m) != 3 {
		return
	}
	c := irc.GetChannel(ev.Server, m[1])
	if c == nil {
		return
	}
	users := strings.Split(strings.TrimRight(m[2], " "), " ", -1)
	for i := 0; i < len(users); i++ {
		u := users[i]
		mode := ""
		if strings.IndexAny(u, "@&~+%") > 0 {
			u = u[1:]
			mode = u[0:1]
		}
		// don't check if the rank already exists in case the
		// map is somehow wrong..
		c.Users[u] = mode
	}
}

// Join a channel with a default configuration
func (irc *Irc) Join(ac *api.Action) {
	conf := ConfigChannel{
		Master: false,
		Name: ac.Channel,
	}
	irc.JoinChannel(conf, ac.Server, ac.Channel)
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

// Called when the bot leaves a channel or is kicked
func (irc *Irc) DestroyChannel(server string, channel string) {
	s := irc.GetServer(server)
	c := irc.GetChannel(server, channel)
	if c != nil {
		c.Destroy <- 42
		<- c.Destroy
		cname := c.Config.Name
		s.Channels[cname] = s.Channels[cname], false
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
		irc.DestroyChannel(ac.Server, ac.Channel)
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
