package main

import (
	"botapi"
	"fmt"
	"strings"
)

func main() {
	config := NewConfig("config.json")
	chac, chev := botapi.ImportFrom(config.RobotInterface, config.ModuleName)

	// action which will be sent to the robot
	a := botapi.Action{
		Type: botapi.A_SAY,
		Priority: botapi.PRIORITY_LOW,
	}

	for {
		e := <- chev
		// if the event is a message !hej, reply by sending an action
		if e.Type == botapi.E_PRIVMSG && len(e.Channel) == 0 {
			for server, targets := range config.Targets {
				a.Server = server
				a.Channel = ""
				a.User = ""
				for _, target := range targets {
					if strings.Index(target, "#") == 1 {
						a.Channel = target
					} else {
						a.User = target
					}
					a.Data = fmt.Sprintf("broadcast> %s: %s", e.User, e.Data)
					chac <- a
				}
			}
		}
	}
}
