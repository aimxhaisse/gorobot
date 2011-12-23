package scripts
// Package scripts implements a module to run shell commands
//
// Shell module can open the administration to send IRC commands using
// the following format:
//
// SERVER PRIORITY IRC_COMMAND
// Where:
//
// -> server is the alias to the server (ie: freenode)
// -> priority is an integer indicating the priority of the command
//    if it's a message (ie: 1, 2, or 3 ; higher is better)
// -> irc_command is a raw IRC command
//
// examples of usages within shelll scripts:
// echo "freenode 1 PRIVMSG aimxhaisse :kenavo" | nc -q 0 localhost $port > /dev/null

import (
	"api"
	"log"
	"net"
	"strconv"
	"strings"
)

// creates a new api.action from what was sent on the admin port
func netAdminCraftAction(output string) api.Action {
	var a api.Action
	shellapi := strings.Split(output, " ")
	a.Type = api.A_RAW
	if len(shellapi) == 3 {
		a.Server = shellapi[0]
		a.Priority, _ = strconv.Atoi(shellapi[1])
		if a.Priority != api.PRIORITY_LOW &&
			a.Priority != api.PRIORITY_MEDIUM &&
			a.Priority != api.PRIORITY_HIGH {
			a.Priority = api.PRIORITY_LOW
		}
		a.Data = shellapi[2]
	} else {
		a.Data = output
		a.Priority = api.PRIORITY_LOW
	}
	return a
}

// shell commands can send several commands in the same connection
// (using \r\n)
func netAdminReadFromCon(con *net.TCPConn, chac chan api.Action) {
	const NBUF = 512
	var rawcmd []byte
	var buf [NBUF]byte

	for {
		n, err := con.Read(buf[0:])
		rawcmd = append(rawcmd, buf[0:n]...)
		if err != nil {
			break
		}
	}
	con.Close()
	msgs := strings.Split(string(rawcmd), "\n")
	for i := 0; i < len(msgs); i++ {
		if len(msgs[i]) > 0 {
			s := strings.TrimRight(msgs[i], " \r\n\t")
			chac <- netAdminCraftAction(s)
		}
	}
}

// open the admin port and directly send RAW commands to the michel
func netAdmin(config Config, chac chan api.Action) {
	a, err := net.ResolveTCPAddr("tcp", "localhost:"+config.LocalPort)
	if err != nil {
		log.Fatalf("Can't resolve to localhost: %v\n", err)
	}
	listener, err := net.ListenTCP("tcp", a)
	if err != nil {
		log.Fatalf("Can't open admin port: %v\n", err)
	}
	for {
		con, err := listener.AcceptTCP()
		if err == nil {
			netAdminReadFromCon(con, chac)
		}
	}
}
