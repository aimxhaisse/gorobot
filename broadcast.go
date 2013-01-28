package main

import (
	"fmt"
	"strings"
)

// Configuration of the broadcast module
type BroadcastConfig struct {
	Targets map[string][]string // List of servers/channels-users to broadcast
}

// Broadcast listens for private messages and broadcasts them to a list of targets
func Broadcast(chac chan Action, chev chan Event, config BroadcastConfig) {
	a := Action{
		Type:     A_SAY,
		Priority: PRIORITY_LOW,
	}
	for {
		e := <-chev
		if e.Type == E_PRIVMSG && len(e.Channel) == 0 {
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
					a.Data = fmt.Sprintf("private> %s: %s", e.User, e.Data)
					chac <- a
				}
			}
		}
	}
}
