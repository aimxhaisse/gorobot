package main

import (
	"json"
	"os"
	"io/ioutil"
	"fmt"
)

type Config struct {
	ModuleName	string
	AdminScripts	string
	PublicScripts	string
	PrivateScripts	string
	LocalPort	string
	RobotInterface	string
}

// Returns a new configuration from file pointed by path
func NewConfig(path string) (*Config) {
	file, e := ioutil.ReadFile(path)
	if e != nil {
		fmt.Printf("Configuration error: %v\n", e)
		os.Exit(1)
	}
	var config Config
	err := json.Unmarshal(file, &config)
	if err != nil {
		fmt.Printf("Configuration error: %s\n", err)
		os.Exit(1)
	}
	return &config
}
