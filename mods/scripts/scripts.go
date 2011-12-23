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
	"api"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// avoid characters such as "../" to disallow commands like "!../admin/kick"
var re_cmd = regexp.MustCompile("^!([a-zA-Z0-9]+)( .*)?")

func fileExists(cmd string) bool {
	_, err := os.Stat(cmd)
	return err == nil
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

	in_params := strings.Split(ev.Data, " ")

	command := exec.Command(path,
		config.LocalPort,
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

func Scripts(chac chan api.Action, chev chan api.Event, config Config) {
	go netAdmin(config, chac)
	for {
		e, ok := <-chev

		if !ok {
			log.Printf("Channel closed")
			return
		}

		switch e.Type {
		case api.E_PRIVMSG:
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
