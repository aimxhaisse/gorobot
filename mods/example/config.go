package example

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	ModuleName     string
	HelloWorld     string
	RobotInterface string
}

// NewConfig returns a new configuration from file pointed by path
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
