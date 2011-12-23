package broadcast

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
		log.Fatalf("Configuration error: %v\n", e)
	}
	var config Config
	err := json.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("Configuration error: %s\n", err)
	}
	return &config
}
