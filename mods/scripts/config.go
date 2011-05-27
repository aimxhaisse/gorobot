package scripts

import (
	"json"
	"io/ioutil"
	"log"
)

type Config struct {
	ModuleName     string
	AdminScripts   string
	PublicScripts  string
	PrivateScripts string
	LocalPort      string
	RobotInterface string
}

// Returns a new configuration from file pointed by path
func NewConfig(path string) *Config {
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
