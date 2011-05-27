// Package gorobot/broadcast implements a gorobot module to broacast private conversations
package broadcast

import (
	"gorobot/api"
	"fmt"
	"strings"
)

// Broadcast listens for private messages and broadcasts them to a list of targets
func Broadcast(chev chan api.Event, chac chan api.Action, config Config) {
	a := api.Action{
		Type:     api.A_SAY,
		Priority: api.PRIORITY_LOW,
	}
	for {
		e := <-chev
		if e.Type == api.E_PRIVMSG && len(e.Channel) == 0 {
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
