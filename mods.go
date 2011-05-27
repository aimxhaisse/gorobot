package mods

import (
	"gorobot/api"
	"gorobot/scripts"
	"gorobot/radio"
	"io/ioutil"
	"log"
	"json"
	"time"
)

type Config struct {
	Scripts		scripts.Config
	Radio		radio.Config
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
	config := newConfig("./mods.json")

	// module scripts
	go func() {
		chac, chev := api.ImportFrom(config.Scripts.RobotInterface, config.Scripts.ModuleName)
		scripts.Scripts(chac, chev, config.Scripts);
	}()

	// module radio
	go func() {
		chac, chev := api.ImportFrom(config.Radio.RobotInterface, config.Radio.ModuleName)
		scripts.Scripts(chac, chev, config.Radio);
	}()

	for {
		time.Sleep(1.e9)
	}
}
