// package scripts is able to associate an IRC command with a script.
// !xyz will execute the script located in public/xyz.cmd
// if public/xyz.cmd doesn't exist and the command was issued by and admin,
// then admin/xyz.cmd is executed.
// if the command was issued in private, then private/xyz.cmd is executed
//
// scripts can open the admin port and send IRC commands directly to the
// server.
package scripts

import (
	"gorobot/api"
	"regexp"
	"exec"
	"os"
	"fmt"
	"strings"
	"log"
)

// avoid characters such as "../" to disallow commands like "!../admin/kick"
var re_cmd = regexp.MustCompile("^!([a-zA-Z0-9]+)( .*)?")

func fileExists(cmd string) bool {
	stat, err := os.Stat(cmd)
	if err == nil {
		return stat.IsRegular()
	}
	return false
}

func cmdPath(config Config, cmd string, admin bool, private bool) string {
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

func execCmd(config Config, path string, ev api.Event) {
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

func Scripts(chac chan api.Action, chev chan api.Event, config Config) {
	go netAdmin(config, chac)
	for {
		log.Printf("Loop")
		e, ok := <-chev

		if !ok {
			log.Printf("Channel closed")
			return
		}

		switch e.Type {
		case api.E_PRIVMSG:
			log.Printf("Got a message")
			if m := re_cmd.FindStringSubmatch(e.Data); len(m) > 0 {
				path := cmdPath(config, m[1],
					e.AdminCmd,
					len(e.Channel) == 0)
				if len(path) > 0 {
					go execCmd(config, path, e)
				}
			}
		}
	}
}
