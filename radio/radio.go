package main

import (
	"fmt"
	"github.com/jteeuwen/go-pkg-mpd"
	"log"
	"time"
	"botapi"
)

func IRCSay(msg string, chac chan botapi.Action, config *Config) {
	for server, channel := range config.Broadcast {
		ac := botapi.Action{
			Data:     msg,
			Priority: botapi.PRIORITY_LOW,
			Type:     botapi.A_SAY,
			Server:   server,
			Channel:  channel,
		}
		chac <- ac
	}
}

func MPDWatch(client *mpd.Client, chac chan botapi.Action, config *Config) {
	str := ""
	prev := str

	for {
		current, err := client.Current()

		if err != nil {
			return
		}

		if len(current["Artist"]) == 0 {
			current["Artist"] = "unknown"
		}
		if len(current["Title"]) > 0 {
			str = fmt.Sprintf("radio m1ch3l: now playing \"%s\" (%s)", current["Title"], current["Artist"])
			if str != prev {
				go IRCSay(str, chac, config)
				prev = str
			}
		}

		time.Sleep(15 * 1e9)
	}
}

func main() {
	config := NewConfig("./mod-radio.json")
	chac, chev := botapi.ImportFrom(config.RobotInterface, config.ModuleName)

	go func(chac chan botapi.Action, config *Config) {
		for {
			log.Printf("Connecting to MPD server")
			client, err := mpd.Dial(config.MPDServer, config.MPDPassword)
			if err == nil {
				log.Printf("MPD: connected")
				MPDWatch(client, chac, config)
			}
			log.Printf("Disconnected from MPD server, retrying in 15 seconds")
			time.Sleep(15 * 1e9)
		}
	}(chac, config)

	for {
		<-chev
	}
}
