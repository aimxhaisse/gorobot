package radio

import (
	"json"
	"log"
	"io/ioutil"
)

type Config struct {
	ModuleName     string            // Name of the module
	RobotInterface string            // Address of the gorobot
	MPDServer      string            // Address of the MPD server
	MPDPassword    string            // Password of the MPD server
	Broadcast      map[string]string // Map of server/channel-users to broadcast music stream
}

// Returns a new Config from the file pointed by path
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
