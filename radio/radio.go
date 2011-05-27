// package radio gorobot/radio implements a gorobot module to broadcast the activity
// of a MDP server
package radio

import (
	"github.com/aimxhaisse/gorobot/api"
	"github.com/jteeuwen/go-pkg-mpd"
	"fmt"
	"log"
	"time"
)

func ircSay(msg string, chac chan api.Action, config Config) {
	for server, channel := range config.Broadcast {
		ac := api.Action{
			Data:     msg,
			Priority: api.PRIORITY_LOW,
			Type:     api.A_SAY,
			Server:   server,
			Channel:  channel,
		}
		chac <- ac
	}
}

func mpdWatch(client *mpd.Client, chac chan api.Action, config Config) {
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
				go ircSay(str, chac, config)
				prev = str
			}
		}

		time.Sleep(15 * 1e9)
	}
}

// Radio watches the activity of a MPD server and broadcast changes
func Radio(chac chan api.Action, chev chan api.Event, config Config) {
	go func(chac chan api.Action, config Config) {
		for {
			log.Printf("Connecting to MPD server")
			client, err := mpd.Dial(config.MPDServer, config.MPDPassword)
			if err == nil {
				log.Printf("MPD: connected")
				mpdWatch(client, chac, config)
			}
			log.Printf("Disconnected from MPD server, retrying in 15 seconds")
			time.Sleep(15 * 1e9)
		}
	}(chac, config)

	for {
		<-chev
	}
}
