package main

import (
	"api"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"mods/broadcast"
	"mods/radio"
	"mods/scripts"
	"os"
	"time"
)

var cfg *string = flag.String("config", "grocket.json", "path to a json configuration file")
var newcfg *bool = flag.Bool("generate-config", false, "generate a default configuration file")

type Config struct {
	Scripts   scripts.Config
	Radio     radio.Config
	Broadcast broadcast.Config
}

// Returns a new configuration from file pointed by path
func NewConfig(path string) *Config {
	file, e := ioutil.ReadFile(path)
	if e != nil {
		log.Fatalf("Configuration error: %v\n", e)
	}
	var config Config
	err := json.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("Configuration error: %s\n", err)
	}
	return &config
}

// Returns  a default configuration
func DefaultConfig() *Config {
	cfg := Config{
		Scripts: scripts.Config{
			ModuleName:     "mod-scripts",
			AdminScripts:   "scripts/admin",
			PublicScripts:  "scripts/public",
			PrivateScripts: "scripts/private",
			LocalPort:      "3112",
			RobotInterface: "localhost:3111",
		},
		Radio: radio.Config{
			ModuleName:     "mod-radio",
			RobotInterface: "localhost:3111",
			MPDServer:      "localhost:6600",
			MPDPassword:    "",
			Broadcast:      make(map[string]string),
		},
		Broadcast: broadcast.Config{
			ModuleName:     "mod-broadcast",
			RobotInterface: "localhost:3111",
			Targets:        make(map[string][]string),
		},
	}
	cfg.Radio.Broadcast["freenode"] = "#sbrk"
	cfg.Broadcast.Targets["freenode"] = []string{"#sbrk"}
	return &cfg
}

func main() {
	flag.Parse()

	if *newcfg == true {
		file, err := os.Create(*cfg)
		if err != nil {
			log.Fatalf("Can't create configuration file: %v", err)
		}
		config := DefaultConfig()
		data, err := json.MarshalIndent(*config, "", " ")
		if err != nil {
			log.Fatalf("Can't create configuration file: %v", err)
		}
		file.Write(data)
		file.Close()
	} else {
		config := NewConfig(*cfg)

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
}
