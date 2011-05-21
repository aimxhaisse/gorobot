package main

// Shell module can open the administration to send IRC commands using
// the following format:

// SERVER PRIORITY IRC_COMMAND

// Where:
// -> server is the alias to the server (ie: freenode)
// -> priority is an integer indicating the priority of the command
//    if it's a message (ie: 1, 2, or 3 ; higher is better)
// -> irc_command is a raw IRC command

// examples of usages within shelll scripts:
// echo "freenode 1 PRIVMSG aimxhaisse :kenavo" | nc -q 0 localhost $port > /dev/null

import (
	"botapi"
	"log"
	"net"
	"strings"
	"strconv"
)

// creates a new botapi.action from what was sent on the admin port
func NetAdminCraftAction(output string) botapi.Action {
	var a botapi.Action
	shellapi := strings.Split(output, " ", 3)
	a.Type = botapi.A_RAW
	if len(shellapi) == 3 {
		a.Server = shellapi[0]
		a.Priority, _ = strconv.Atoi(shellapi[1])
		if a.Priority != botapi.PRIORITY_LOW &&
			a.Priority != botapi.PRIORITY_MEDIUM &&
			a.Priority != botapi.PRIORITY_HIGH {
			a.Priority = botapi.PRIORITY_LOW
		}
		a.Data = shellapi[2]
	} else {
		a.Data = output
		a.Priority = botapi.PRIORITY_LOW
	}
	return a
}

// shell commands can send several commands in the same connection
// (using \r\n)
func NetAdminReadFromCon(con *net.TCPConn, chac chan botapi.Action) {
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
	msgs := strings.Split(string(rawcmd), "\n", -1)
	for i := 0; i < len(msgs); i++ {
		if len(msgs[i]) > 0 {
			s := strings.TrimRight(msgs[i], " \r\n\t")
			chac <- NetAdminCraftAction(s)
		}
	}
}

// open the admin port and directly send RAW commands to the michel
func NetAdmin(config Config, chac chan botapi.Action) {
	a, err := net.ResolveTCPAddr("tcp", "localhost:"+config.LocalPort)
	if err != nil {
		log.Panic("Can't resolve to localhost\n")
	}
	listener, err := net.ListenTCP("tcp", a)
	if err != nil {
		log.Panic("Can't open admin port\n")
	}
	for {
		con, err := listener.AcceptTCP()
		if err == nil {
			NetAdminReadFromCon(con, chac)
		}
	}
}
