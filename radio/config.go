package main

import (
	"json"
	"log"
	"io/ioutil"
)

type Config struct {
	ModuleName	string
	RobotInterface	string
	MPDServer	string
	MPDPassword	string
	Broadcast	map[string] string
}

// Returns a new configuration from file pointed by path
func NewConfig(path string) (*Config) {
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
