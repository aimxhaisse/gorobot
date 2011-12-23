// package rocket launches some of the available modules of gorobot
// feel free to modify it to your needs
package rocket

import (
	"api"
	"encoding/json"
	"io/ioutil"
	"log"
	"mods/broadcast"
	"mods/radio"
	"mods/scripts"
	"time"
)

type Config struct {
	Scripts   scripts.Config
	Radio     radio.Config
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
