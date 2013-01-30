package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// Configuration of the stats module
type StatsConfig struct {
	Timeout    time.Duration // Timeout in seconds
	Server     string        // Server to monitor
	Channel    string        // Chan to monitor
	OutputFile string        // File to write logs to
}

// Stats outut regularly outputs the number of users on a channel to an output file
func Stats(chac chan Action, chev chan Event, config StatsConfig) {
	var nb_users int
	file, err := os.OpenFile(config.OutputFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("can't create stats file: %v", err)
		return
	}
	for {
		select {
		case <-time.NewTicker(config.Timeout * time.Second).C:
			var a Action
			a.Server = config.Server
			a.Channel = config.Channel
			a.Priority = PRIORITY_HIGH
			a.Type = A_NAMES
			chac <- a

		case ev := <-chev:
			if ev.Type == E_NAMES {
				nb_users = nb_users + len(strings.Fields(ev.Data))
			} else if ev.Type == E_ENDOFNAMES {
				file.WriteString(fmt.Sprintf("%s - %d users\n", time.Now().String(), nb_users))
				nb_users = 0
			}
		}
	}
}
