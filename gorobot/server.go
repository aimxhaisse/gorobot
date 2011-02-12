package gorobot

import (
	"net"
	"botapi"
	"log"
	"fmt"
	"bufio"
	"time"
	"os"
	"strings"
)

// IRC Server
type Server struct {
	Config		ConfigServer
	SendMeRaw	map[int] chan string	// Channel to send raw commands to the server
	Socket		net.Conn		// Socket to the server
	Destroy		chan int		// Writing here will destroy the server
}

// Creates a new connection to a server
func NewServer(conf *ConfigServer, chev chan botapi.Event) *Server {
	log.Printf("connecting to [%s]\n", conf.Host)
	connection, err := net.Dial("tcp", "", conf.Host)
	if err != nil {
		log.Printf("can't connect to %s", conf.Host)
		return nil
	}
	serv := Server{
		Config: *conf,
		SendMeRaw: make(map[int] chan string),
		Socket: connection,
		Destroy: make(chan int),
	}
	serv.SendMeRaw[botapi.PRIORITY_LOW] = make(chan string)
	serv.SendMeRaw[botapi.PRIORITY_MEDIUM] = make(chan string)
	serv.SendMeRaw[botapi.PRIORITY_HIGH] = make(chan string)
	serv.Init(chev)
	return &serv
}

// Initialize a new connection to the server
func (serv *Server) Init(chev chan botapi.Event) {
	go reader(serv.Config.Name, serv.Socket, chev)
	go writer(serv.Socket, serv.SendMeRaw, serv.Destroy)
	serv.SendMeRaw[botapi.PRIORITY_HIGH] <-	fmt.Sprintf("NICK %s\r\n", serv.Config.Nickname)
	serv.SendMeRaw[botapi.PRIORITY_HIGH] <-	fmt.Sprintf("USER %s 0.0.0.0 0.0.0.0 :%s\r\n", serv.Config.Username, serv.Config.Realname)
}

func (serv *Server) Say(ac *botapi.Action) {
	if len(ac.Channel) > 0 {
		serv.SendMeRaw[ac.Priority] <- fmt.Sprintf("PRIVMSG %s :%s\r\n", ac.Channel, ac.Data)
	} else {
		serv.SendMeRaw[ac.Priority] <- fmt.Sprintf("PRIVMSG %s :%s\r\n", ac.User, ac.Data)
	}
}

func (serv *Server) LeaveChannel(name string, msg string) {
	if len(msg) > 0 {
		serv.SendMeRaw[botapi.PRIORITY_HIGH] <- fmt.Sprintf("PART %s :%s\r\n", name, msg)
	} else {
		serv.SendMeRaw[botapi.PRIORITY_HIGH] <- fmt.Sprintf("PART %s\r\n", name, msg)
	}
}

func (serv *Server) KickUser(channel string, user string, msg string) {
	if len(msg) > 0 {
		serv.SendMeRaw[botapi.PRIORITY_HIGH] <- fmt.Sprintf("KICK %s %s :%s\r\n", channel, user, msg)
 	} else {
		serv.SendMeRaw[botapi.PRIORITY_HIGH] <- fmt.Sprintf("KICK %s %s\r\n", channel, user)
	}
}

func (serv *Server) JoinChannel(name string) {
	var ok bool
	var conf *ConfigChannel

	if conf, ok = serv.Config.Channels[name]; ok == false {
		conf.Master = false
		conf.Name = name
		conf.Password = ""
	}

	if len(conf.Password) > 0 {
		serv.SendMeRaw[botapi.PRIORITY_HIGH] <- fmt.Sprintf("JOIN %s %s\r\n", conf.Name, conf.Password)
	} else {
		serv.SendMeRaw[botapi.PRIORITY_HIGH] <- fmt.Sprintf("JOIN %s\r\n", conf.Name)
	}
}

// Extract events from the server
func reader(serv_name string, connection net.Conn, chev chan botapi.Event) {
	r := bufio.NewReader(connection)
	for {
		var err os.Error
		var p []byte
		if p, err = r.ReadSlice('\n'); err != nil {
			return
		}
		line := strings.TrimRight(string(p), "\r\t\n")
		ev := ExtractEvent(line)
		if ev != nil {
			ev.Server = serv_name
			log.Printf("\x1b[1;36m%s\x1b[0m", line)
			chev <- *ev
		}
	}
}

// Send the raw command to the server
func writerSend(after *int64, ahead *int64, before *int64, str string, connection *net.Conn) {
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
	raw := []byte(str)
	log.Printf("\x1b[1;35m%s\x1b[0m", strings.TrimRight(str, "\r\t\n"))
	if _, err := (*connection).Write(raw); err != nil {
		// @todo destroy.
		return
	}
	*ahead += 2e9
	*before = time.Nanoseconds()
}

// Pick raw commands in order of priority
func writer(connection net.Conn, chin map[int] chan string, destroy chan int) {
	var after int64 = 0
	var ahead int64 = 0
	before := time.Nanoseconds()

	for {
		select {
		case <- destroy:
			destroy <- 42
			return
		case str := <- chin[botapi.PRIORITY_HIGH]:
			writerSend(&after, &ahead, &before, str, &connection)
		case str := <- chin[botapi.PRIORITY_MEDIUM]:
			writerSend(&after, &ahead, &before, str, &connection)
		case str := <- chin[botapi.PRIORITY_LOW]:
			writerSend(&after, &ahead, &before, str, &connection)
		}
	}
}
