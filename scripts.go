package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type CommandLogger interface {
	LogCommand(server string, channel string, from string, cmd string)
}

type ScriptsConfig struct {
	AdminScripts   string
	PublicScripts  string
	PrivateScripts string
	LocalPort      string
}

// avoid characters such as "../" to disallow commands like "!../admin/kick"
var re_cmd = regexp.MustCompile("^!([a-zA-Z0-9]+)( .*)?")

func fileExists(cmd string) bool {
	_, err := os.Stat(cmd)
	return err == nil
}

func cmdPath(config ScriptsConfig, cmd string, admin bool, private bool) string {
	if private {
		path := fmt.Sprintf("%s/%s.cmd", config.PrivateScripts, cmd)
		if fileExists(path) {
			return path
		}
		return ""
	}
	if admin {
		path := fmt.Sprintf("%s/%s.cmd", config.AdminScripts, cmd)
		if fileExists(path) {
			return path
		}
	}
	path := fmt.Sprintf("%s/%s.cmd", config.PublicScripts, cmd)
	if fileExists(path) {
		return path
	}
	return ""
}

func execCmd(config ScriptsConfig, path string, ev Event) {
	log.Printf("Executing [%s]\n", path)

	in_params := strings.Split(ev.Data, " ")
	dynamic_hostname := strings.Split(config.LocalPort, ":")

	command := exec.Command(path,
		dynamic_hostname[1],
		ev.Server,
		ev.Channel,
		ev.User)

	for _, v := range in_params[1:] {
		command.Args = append(command.Args, v)
	}

	err := command.Run()
	if err == nil {
		command.Wait()
	}
}

// creates a new action from what was sent on the admin port
func netAdminCraftAction(output string) Action {
	var a Action
	shellapi := strings.SplitN(output, " ", 3)
	a.Type = A_RAW
	if len(shellapi) == 3 {
		a.Server = shellapi[0]
		a.Priority, _ = strconv.Atoi(shellapi[1])
		if a.Priority != PRIORITY_LOW &&
			a.Priority != PRIORITY_MEDIUM &&
			a.Priority != PRIORITY_HIGH {
			a.Priority = PRIORITY_LOW
		}
		a.Data = shellapi[2]
	} else {
		a.Data = output
		a.Priority = PRIORITY_LOW
	}
	return a
}

// shell commands can send several commands in the same connection
// (using \r\n)
func netAdminReadFromCon(con *net.TCPConn, chac chan Action) {
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

// opens the admin port and directly send RAW commands to grobot
func netAdmin(config ScriptsConfig, chac chan Action) {
	a, err := net.ResolveTCPAddr("tcp", config.LocalPort)
	if err != nil {
		log.Fatalf("Can't resolve: %v\n", err)
	}
	listener, err := net.ListenTCP("tcp", a)
	if err != nil {
		log.Fatalf("Can't open admin port: %v\n", err)
	}
	for {
		con, err := listener.AcceptTCP()
		if err == nil {
			go netAdminReadFromCon(con, chac)
		}
	}
}

func Scripts(chac chan Action, chev chan Event, logger CommandLogger, config ScriptsConfig) {
	go netAdmin(config, chac)
	for {
		e, ok := <-chev

		if !ok {
			log.Printf("Channel closed")
			return
		}

		switch e.Type {
		case E_PRIVMSG:
			if m := re_cmd.FindStringSubmatch(e.Data); len(m) > 0 {
				path := cmdPath(config, m[1],
					e.AdminCmd,
					len(e.Channel) == 0)
				if len(path) > 0 {
					logger.LogCommand(e.Server, e.Channel, e.User, m[1])
					go execCmd(config, path, e)
				}
			}
		}
	}
}
