package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// IRC Bot
type Irc struct {
	Events  chan Event         // Events are written here
	Servers map[string]*Server // Servers where the bot is connected
}

// IRC Server
type Server struct {
	Config    ConfigServer        // Configuration of the server
	SendMeRaw map[int]chan string // Channel to send raw commands to the server
	Socket    net.Conn            // Socket to the server
	Connected bool                // Whether we are connected or not to the server
}

// NewServer creates a new IRC server and connects to it
func NewServer(conf *ConfigServer, chev chan Event) *Server {
	log.Printf("connecting to %s (%s)\n", conf.Name, conf.Host)
	serv := Server{
		Config:    *conf,
		SendMeRaw: make(map[int]chan string),
		Connected: false,
	}
	serv.SendMeRaw[PRIORITY_LOW] = make(chan string)
	serv.SendMeRaw[PRIORITY_MEDIUM] = make(chan string)
	serv.SendMeRaw[PRIORITY_HIGH] = make(chan string)
	connection, err := net.Dial("tcp", conf.Host)
	if err != nil {
		log.Printf("can't connect to %s (%s)", conf.Name, conf.Host)
		return &serv
	}
	serv.Socket = connection
	serv.Init(chev, serv.Config.FloodControl)
	return &serv
}

// Init initialized a new connection to the server, and identifies to bot
func (serv *Server) Init(chev chan Event, flood_control bool) {
	destroy := make(chan int)
	go reader(destroy, serv.Config.Name, serv.Socket, chev)
	go writer(destroy, serv.Socket, serv.SendMeRaw, flood_control)
	serv.Connected = true
	if len(serv.Config.Password) > 0 {
		serv.SendRawCommand(fmt.Sprintf("PASS %s\r\n", serv.Config.Password), PRIORITY_HIGH)
	}
	serv.SendRawCommand(fmt.Sprintf("NICK %s\r\n", serv.Config.Nickname), PRIORITY_HIGH)
	serv.SendRawCommand(fmt.Sprintf("USER %s 0.0.0.0 0.0.0.0 :%s\r\n", serv.Config.Username, serv.Config.Realname), PRIORITY_HIGH)
	log.Printf("connected to %s (%s)", serv.Config.Name, serv.Config.Host)
}

// TryReconnect attempts to reconnect to an IRC server which has been disconnected
func (serv *Server) TryReconnect(chev chan Event) {
	log.Printf("trying to reconnect to %s (%s)", serv.Config.Name, serv.Config.Host)
	connection, err := net.Dial("tcp", serv.Config.Host)
	if err != nil {
		log.Printf("can't reconnect to %s (%s)", serv.Config.Name, serv.Config.Host)
		return
	}
	serv.Socket = connection
	serv.Init(chev, serv.Config.FloodControl)
}

// Say sends a message on the server to the specified target (channel or user)
func (serv *Server) Say(ac *Action) {
	if len(ac.Channel) > 0 {
		serv.SendRawCommand(fmt.Sprintf("PRIVMSG %s :%s\r\n", ac.Channel, ac.Data), ac.Priority)
	} else {
		serv.SendRawCommand(fmt.Sprintf("PRIVMSG %s :%s\r\n", ac.User, ac.Data), ac.Priority)
	}
}

func (serv *Server) Names(ac *Action) {
	if len(ac.Channel) > 0 {
		serv.SendRawCommand(fmt.Sprintf("NAMES %s\r\n", ac.Channel), ac.Priority)
	}
}

// Disconnect disconnects from the server
func (serv *Server) Disconnect() {
	log.Printf("disconnected from %s (%s)", serv.Config.Name, serv.Config.Host)
	serv.Connected = false
	serv.Socket.Close()
}

// LeaveChannel leaves the specified channel
func (serv *Server) LeaveChannel(name string, msg string) {
	if len(msg) > 0 {
		serv.SendRawCommand(fmt.Sprintf("PART %s :%s\r\n", name, msg), PRIORITY_HIGH)
	} else {
		serv.SendRawCommand(fmt.Sprintf("PART %s\r\n", name, msg), PRIORITY_HIGH)
	}
}

// KickUser kicks the specified user
func (serv *Server) KickUser(channel string, user string, msg string) {
	if len(msg) > 0 {
		serv.SendRawCommand(fmt.Sprintf("KICK %s %s :%s\r\n", channel, user, msg), PRIORITY_HIGH)
	} else {
		serv.SendRawCommand(fmt.Sprintf("KICK %s %s\r\n", channel, user), PRIORITY_HIGH)
	}
}

// JoinChannel joins the specified channel
func (serv *Server) JoinChannel(name string) {
	var ok bool
	var conf *ConfigChannel

	if conf, ok = serv.Config.Channels[name]; ok == false {
		conf.Master = false
		conf.Name = name
		conf.Password = ""
	}

	if len(conf.Password) > 0 {
		serv.SendRawCommand(fmt.Sprintf("JOIN %s %s\r\n", conf.Name, conf.Password), PRIORITY_HIGH)
	} else {
		serv.SendRawCommand(fmt.Sprintf("JOIN %s\r\n", conf.Name), PRIORITY_HIGH)
	}
}

// SendRawCommand sends a raw IRC command to the server, priority is meaningful when the server has the flood control enabled
func (serv *Server) SendRawCommand(cmd string, priority int) {
	if serv.Connected == true {
		go func(cmd string, priority int) {
			serv.SendMeRaw[priority] <- cmd
		}(cmd, priority)
	}
}

func reader(destroy chan int, serv_name string, connection net.Conn, chev chan Event) {
	r := bufio.NewReader(connection)
	for {
		var line string
		var err error
		if line, err = r.ReadString('\n'); err != nil {
			chev <- Event{
				Server: serv_name,
				Type:   E_DISCONNECT,
			}
			log.Printf("read error on %s: %v", serv_name, err)
			destroy <- 0
			return
		}
		line = strings.TrimRight(line, "\r\t\n")
		ev := ExtractEvent(line)
		if ev != nil {
			ev.Server = serv_name
			log.Printf("\x1b[1;36m%s\x1b[0m", line)
			chev <- *ev
		}
	}
}

func writerSendNoFlood(str string, connection net.Conn) bool {
	raw := []byte(str)
	log.Printf("\x1b[1;35m%s\x1b[0m", strings.TrimRight(str, "\r\t\n"))
	if _, err := connection.Write(raw); err != nil {
		connection.Close()
		log.Printf("can't write on socket: %v", err)
		return false
	}
	return true
}

func writerSendFlood(after *time.Time, ahead *time.Duration, before *time.Time, str string, connection net.Conn) bool {
	// "while the timer is less than ten seconds ahead of the current time, parse any
	// present messages and penalize the client by 2 seconds for each message" (doc irssi)
	*after = time.Now()
	*ahead -= after.Sub(*before)
	if ahead.Seconds() < 0 {
		*ahead = time.Duration(0 * time.Second)
	} else if ahead.Seconds() > 10 {
		time.Sleep(time.Duration((ahead.Seconds() - float64(10))) * time.Second)
		*ahead = time.Duration(10 * time.Second)
	}
	raw := []byte(str)
	log.Printf("\x1b[1;35m%s\x1b[0m", strings.TrimRight(str, "\r\t\n"))
	if _, err := connection.Write(raw); err != nil {
		connection.Close()
		log.Printf("can't write on socket: %v", err)
		return false
	}
	*ahead += 2 * time.Second
	*before = time.Now()
	return true
}

func writer(destroy chan int, connection net.Conn, chin map[int]chan string, flood_control bool) {
	var after time.Time
	var ahead time.Duration

	before := time.Now()
	for {
		select {
		case <-destroy:
			return
		case str := <-chin[PRIORITY_HIGH]:
			if !writerDispatch(&after, &ahead, &before, str, connection, flood_control) {
				return
			}
		case str := <-chin[PRIORITY_MEDIUM]:
			if !writerDispatch(&after, &ahead, &before, str, connection, flood_control) {
				return
			}
		case str := <-chin[PRIORITY_LOW]:
			if !writerDispatch(&after, &ahead, &before, str, connection, flood_control) {
				return
			}
		}
	}
}

func writerDispatch(after *time.Time, ahead *time.Duration, before *time.Time, str string, connection net.Conn, flood_control bool) bool {
	if flood_control {
		return writerSendFlood(after, ahead, before, str, connection)
	}
	return writerSendNoFlood(str, connection)
}

// NewIRC creates a  new IRC bot
func NewIrc() *Irc {
	b := Irc{
		Events:  make(chan Event),
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
