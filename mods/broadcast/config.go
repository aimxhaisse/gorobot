package broadcast

import (
	"json"
	"log"
	"io/ioutil"
)

type Config struct {
	ModuleName     string              // Name of the module registered in the IRC robot
	RobotInterface string              // Interface of the IRC robot to use for connection
	Targets        map[string][]string // List of servers/channels-users to broadcast
}

// NewConfig a new configuration from the file (formatted in JSON) pointed by path
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
