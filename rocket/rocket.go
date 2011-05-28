// package rocket launches some of the available modules of gorobot
// feel free to modify it to your needs
package rocket

import (
	"gorobot/api"
	"gorobot/mods/scripts"
	"gorobot/mods/radio"
	"gorobot/mods/rss"
	"gorobot/mods/broadcast"
	"io/ioutil"
	"log"
	"json"
	"time"
)

type Config struct {
	Scripts   scripts.Config
	Radio     radio.Config
	Rss       rss.Config
	Broadcast broadcast.Config
}

// Returns a new configuration from file pointed by path
func newConfig(path string) *Config {
	file, e := ioutil.ReadFile(path)
	if e != nil {
		log.Panic("Configuration error: %v\n", e)
	}
	var config Config
	err := json.Unmarshal(file, &config)
	if err != nil {
		log.Panic("Configuration error: %s\n", err)
	}
	return &config
}

func main() {
	config := newConfig("./rocket.json")

	// module scripts
	go func() {
		chac, chev := api.ImportFrom(config.Scripts.RobotInterface, config.Scripts.ModuleName)
		scripts.Scripts(chac, chev, config.Scripts)
	}()

	// module radio
	go func() {
		chac, chev := api.ImportFrom(config.Radio.RobotInterface, config.Radio.ModuleName)
		radio.Radio(chac, chev, config.Radio)
	}()

	// module rss
	go func() {
		chac, chev := api.ImportFrom(config.Rss.RobotInterface, config.Rss.ModuleName)
		rss.Rss(chac, chev, config.Rss)
	}()

	// module broadcast
	go func() {
		chac, chev := api.ImportFrom(config.Broadcast.RobotInterface, config.Broadcast.ModuleName)
		broadcast.Broadcast(chac, chev, config.Broadcast)
	}()

	// add you own

	for {
		time.Sleep(1.e9)
	}
}
