package main

import (
	"api"
)

func main() {
	config := NewConfig("config.json")
	chac, chev := api.ImportFrom(config.RobotInterface, config.ModuleName)

	// action which will be sent to the robot
	a := api.Action{
		Type: api.A_SAY,
		Data: config.HelloWorld,
		Priority: api.PRIORITY_LOW,
	}

	for {
		e := <- chev
		// if the event is a message !hej, reply by sending an action
		if e.Type == api.E_PRIVMSG && len(e.Channel) > 0 && e.Data == "!hej" {
			a.Server = e.Server
			a.Channel = e.Channel
			go func (a api.Action) { chac <- a } (a);
		}
	}
}
