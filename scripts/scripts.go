package main

// this module is able to associate a specific command with a script.
// !xyz will execute the script located in public/xyz.cmd
// if public/xyz.cmd doesn't exist and the command was issued by and admin,
// then admin/xyz.cmd is executed.
//
// scripts can behave into two fashions:
// - they can write to stdout, which will write to the channel where
// the cmd was invoked.
// - they can open the admin port and send IRC commands directly to the
// server.

// @todo:
// -> clean the code
// -> when the command is not an executable, the module crashes

import (
	"api"
	"flag"
	"log"
	"regexp"
	"exec"
	"os"
	"fmt"
	"net"
	"strings"
	"strconv"
)

// avoid characters such as "../" to disallow commands like "!../admin/kick"
var re_cmd = regexp.MustCompile("^!([a-zA-Z0-9]+)( .*)?")
var addr = flag.String("s", "", "address:port of exported netchans")
const ADMIN_PORT = "23456"

func CraftActionSay(e api.Event, output string) (api.Action) {
	var a api.Action
	a.Server = e.Server
	a.Channel = e.Channel
	a.User = e.User
	a.Data = output
	a.Type = api.A_SAY
	a.Priority = api.PRIORITY_LOW
	return a
}

// assumes that output follows the following pattern:
// SERVER PRIORITY RAW_CMD
// example: irc.freenode.org PRIVMSG ...
func CraftActionRaw(output string) (api.Action) {
	var a api.Action
	shellapi := strings.Split(output, " ", 3)
	a.Type = api.A_RAW
	if len(shellapi) == 3 {
		a.Server = shellapi[0]
		a.Priority, _ = strconv.Atoi(shellapi[1])
		if	a.Priority != api.PRIORITY_LOW &&
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

func FileExists(cmd string) (bool) {
	stat, err := os.Stat(cmd)
	if err == nil {
		return stat.IsRegular()
	}
	return false
}

func GetCmdPath(cmd string, admin bool) (string) {
	path := "public/" + cmd + ".cmd"
	if FileExists(path) {
		return path
	}
	if admin {
		path := "admin/" + cmd + ".cmd"
		if FileExists(path) {
			return path
		}
	}
	return ""
}

// "!script param1 param2" will result in the following call:
// "./public/script.cmd port server channel user param1 param2"
func ExecCmd(path string, ev api.Event) (string, os.Error) {
	var result []byte
	argv := []string{path, ADMIN_PORT, ev.Server, ev.Channel, ev.User}
	args := strings.Split(ev.Data, " ", 2)
	if len(args) == 2 {
		parameters := strings.Split(args[1], " ", -1)
		argv = append(argv, parameters...)
	}

	cmd, _ := exec.Run(path, argv,
		[]string{}, "", exec.Pipe, exec.Pipe, exec.Pipe)

	const NBUF = 512
	var buf [NBUF]byte
	for {
		n, err := cmd.Stdout.Read(buf[0:])
		result = append(result, buf[0:n]...)
		if err != nil {
			if err == os.EOF {
				break
			}
			return "", err
		}
	}
	cmd.Wait(0)
	cmd.Close()
	return string(result), nil
}

// open the admin port and directly send RAW commands to the michel
func NetAdmin(chac chan api.Action) os.Error {
	listener, err := net.Listen("tcp", "0.0.0.0:" + ADMIN_PORT)
	if err != nil {
		log.Exit("Can't open admin port\n")
		return err
	}
	for {
		con, err := listener.Accept()
		if err == nil {
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
			s := strings.TrimRight(string(rawcmd), "\r\n")
			fmt.Printf("New raw action received (%s)\n", s)
			chac <- CraftActionRaw(s)
		} else {
			fmt.Printf("Can't accept new connection\n")
		}
	}
	return nil
}

func main() {
	flag.Parse()
	if *addr == "" {
		log.Exit("Usage : ./module -s addr:port")
	}
	chac, chev := api.ImportFrom(*addr, "scripts")
	go NetAdmin(chac)
	for {
		e := <- chev
		if e.Type != api.E_PRIVMSG || len(e.Channel) == 0 {
			continue
		}
		if m := re_cmd.FindStringSubmatch(e.Data); len(m) > 0 {
			path := GetCmdPath(m[1], e.AdminCmd)
			if len(path) > 0 {
				go func(path string, e api.Event) {
					fmt.Printf("Executing %s\n", path)
					output, _ := ExecCmd(path, e)
					if len(output) > 0 {
						msgs := strings.Split(output, "\n", -1)
						for i := 0; i < len(msgs); i++ {
							if len(msgs[i]) > 0 {
								chac <- CraftActionSay(e, msgs[i])
							}
						}
					}
					// if no output, this is probably a command that
  					// will use the administration port (ADMIN_PORT)
					// to send a raw command.
				}(path, e)
			}
		}
	}
}
