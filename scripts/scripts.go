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

// @todo handle Private commands
func GetCmdPath(config *Config, cmd string, admin bool) (string) {
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

func ExecCmd(config Config, path string, ev api.Event) (string, os.Error) {
	var result []byte
	argv := []string{path, config.LocalPort, ev.Server, ev.Channel, ev.User}
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

func TreatEventCmd(config Config, path string, e api.Event, chac chan api.Action) {
	fmt.Printf("Executing %s\n", path)
	output, _ := ExecCmd(config, path, e)
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
}

func main() {
	config := NewConfig("config.json")
	chac, chev := api.ImportFrom(config.RobotInterface, config.ModuleName)
	go NetAdmin(*config, chac)

	for {
		e := <- chev
		if e.Type != api.E_PRIVMSG || len(e.Channel) == 0 {
			continue
		}
		if m := re_cmd.FindStringSubmatch(e.Data); len(m) > 0 {
			path := GetCmdPath(config, m[1], e.AdminCmd)
			if len(path) > 0 {
				go TreatEventCmd(*config, path, e, chac)
			}
		}
	}
}
