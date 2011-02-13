package main

// this module is able to associate a specific command with a script.
// !xyz will execute the script located in public/xyz.cmd
// if public/xyz.cmd doesn't exist and the command was issued by and admin,
// then admin/xyz.cmd is executed.
// if the command was issued in private, then private/xyz.cmd is executed
//
// scripts can open the admin port and send IRC commands directly to the
// server.

import (
	"botapi"
	"regexp"
	"exec"
	"os"
	"fmt"
	"strings"
	"log"
	"time"
)

// avoid characters such as "../" to disallow commands like "!../admin/kick"
var re_cmd = regexp.MustCompile("^!([a-zA-Z0-9]+)( .*)?")

func CraftActionSay(e botapi.Event, output string) (botapi.Action) {
	var a botapi.Action
	a.Server = e.Server
	a.Channel = e.Channel
	a.User = e.User
	a.Data = output
	a.Type = botapi.A_SAY
	a.Priority = botapi.PRIORITY_LOW
	return a
}

func FileExists(cmd string) (bool) {
	stat, err := os.Stat(cmd)
	if err == nil {
		return stat.IsRegular()
	}
	return false
}

func GetCmdPath(config *Config, cmd string, admin bool, private bool, auth bool) (string) {
	if private {
		path := fmt.Sprintf("%s/%s.cmd", config.PrivateScripts, cmd)
		if FileExists(path) {
			return path
		}
		return ""
	}
	path := fmt.Sprintf("%s/%s.cmd", config.PublicScripts, cmd)
	if FileExists(path) {
		return path
	}
	if auth {
		path := fmt.Sprintf("%s/%s.cmd", config.AuthenticatedScripts, cmd)
		if FileExists(path) {
			return path
		}
	}
	if admin {
		path := fmt.Sprintf("%s/%s.cmd", config.AdminScripts, cmd)
		if FileExists(path) {
			return path
		}
	}
	return ""
}

func ExecCmd(config Config, path string, ev botapi.Event) {
	log.Printf("Executing [%s]\n", path)
	argv := []string{path, config.LocalPort, ev.Server, ev.Channel, ev.User}
	args := strings.Split(ev.Data, " ", 2)
	if len(args) == 2 {
		parameters := strings.Split(args[1], " ", -1)
		argv = append(argv, parameters...)
	}
	cmd, err := exec.Run(path, argv,
		[]string{}, "", exec.Pipe, exec.Pipe, exec.Pipe)
	if err == nil {
		cmd.Wait(0)
	}
}

func main() {
	config := NewConfig("config.json")
	chac, chev := botapi.ImportFrom(config.RobotInterface, config.ModuleName)
	users := NewUsers()

	users.Refresh(config.UsersDatabase)
	lastRefresh := time.Seconds()

	go NetAdmin(*config, chac)

	for {
		e := <- chev

		if ((time.Seconds() - lastRefresh) > config.UsersRefreshTimeout) {
			lastRefresh = time.Seconds()
			users.Refresh(config.UsersDatabase)
		}

		switch e.Type {
		case botapi.E_PART:
			users.Logout(e.User)
		case botapi.E_QUIT:
			users.Logout(e.User)
		case botapi.E_KICK:
			users.Logout(e.Data)
		case botapi.E_NICK:
			users.Rename(e.User, e.Data)
		case botapi.E_PRIVMSG:
			if m := re_cmd.FindStringSubmatch(e.Data); len(m) > 0 {
				if m[1] == "identify" && len(m) == 3 {
					args := strings.Split(m[2], " ", -1)
					if len(args) == 3 {
						users.Login(e.User, args[1], args[2])
					}
				}
				path := GetCmdPath(config, m[1],
					e.AdminCmd,
					len(e.Channel) == 0,
					users.IsAuthenticated(e.User))
				if len(path) > 0 {
					go ExecCmd(*config, path, e)
				}
			}
		}
	}
}
