package main

import (
	"botapi"
)

func main() {
	config := NewConfig("./mod-example.json")
	chac, chev := botapi.ImportFrom(config.RobotInterface, config.ModuleName)

	// action which will be sent to the robot
	a := botapi.Action{
		Type:     botapi.A_SAY,
		Data:     config.HelloWorld,
		Priority: botapi.PRIORITY_LOW,
	}

	for {
		e := <-chev
		// if the event is a message !hej, reply by sending an action
		if e.Type == botapi.E_PRIVMSG && len(e.Channel) > 0 && e.Data == "!hej" {
			a.Server = e.Server
			a.Channel = e.Channel
			go func(a botapi.Action) { chac <- a }(a)
		}
	}
}
