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
	"api"
	"regexp"
	"exec"
	"os"
	"fmt"
	"strings"
)

// avoid characters such as "../" to disallow commands like "!../admin/kick"
var re_cmd = regexp.MustCompile("^!([a-zA-Z0-9]+)( .*)?")

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

func FileExists(cmd string) (bool) {
	stat, err := os.Stat(cmd)
	if err == nil {
		return stat.IsRegular()
	}
	return false
}

func GetCmdPath(config *Config, cmd string, admin bool, private bool) (string) {
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
	if admin {
		path := fmt.Sprintf("%s/%s.cmd", config.AdminScripts, cmd)
		if FileExists(path) {
			return path
		}
	}
	return ""
}

func ExecCmd(config Config, path string, ev api.Event) {
	fmt.Printf("Executing [%s]\n", path)
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
	chac, chev := api.ImportFrom(config.RobotInterface, config.ModuleName)
	go NetAdmin(*config, chac)

	for {
		e := <- chev

		if e.Type != api.E_PRIVMSG {
			continue
		}
		if m := re_cmd.FindStringSubmatch(e.Data); len(m) > 0 {
			path := GetCmdPath(config, m[1], e.AdminCmd, len(e.Channel) == 0)
			if len(path) > 0 {
				go ExecCmd(*config, path, e)
			}
		}
	}
}
