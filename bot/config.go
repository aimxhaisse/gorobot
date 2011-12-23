package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	AutoRejoinOnKick bool                     // rejoin channel when kicked
	CronTimeout      int64                    // timeout to perform cron actions such as reconnecting to disconnected servers
	AutoRunModules   bool                     // executes a list of modules at startup
	Logs             ConfigLogs               // log configuration
	Module           ConfigModule             // module configuration
	Servers          map[string]*ConfigServer // server to connects to
}

type ConfigLogs struct {
	Enable            bool   // enable logging
	Directory         string // directory to store logs
	RecordEvents      bool   // record events
	RecordMemoryUsage bool   // record memory usage
	RecordStatistics  bool   // record statistics (not implemented)
}

type ConfigModule struct {
	Interface      string   // interface to be bound to
	AutoRunModules bool     // executes a list of modules at startup
	AutoRun        []string // list of modules to execute
}

type ConfigServer struct {
	Name         string                    // alias of the server
	Host         string                    // address of the server
	FloodControl bool                      // enable flood control
	Nickname     string                    // nickname of the IRC robot
	Realname     string                    // real name of the IRC robot
	Username     string                    // username of the IRC robot
	Password     string                    // password of the IRC server
	Channels     map[string]*ConfigChannel // list of channels to join
}

type ConfigChannel struct {
	Name     string // name of the channel
	Password string // password of the channel
	Master   bool   // enables admin commands on that channel
}

// Returns a new configuration from file pointed by path
func NewConfig(path string) *Config {
	file, e := ioutil.ReadFile(path)
	if e != nil {
		log.Panic("Configuration error: %s\n", e)
	}
	var config Config
	e = json.Unmarshal(file, &config)
	if e != nil {
		log.Panic("Configuration error: %s\n", e)
	}

	for kserv, serv := range config.Servers {
		for kchannel, _ := range serv.Channels {
			config.Servers[kserv].Channels[kchannel].Name = kchannel
		}
	}

	return &config
}

// Returns a default configuration file
func DefaultConfig() *Config {
	var config Config

	config.AutoRejoinOnKick = true
	config.CronTimeout = 60
	config.AutoRunModules = false

	config.Logs.Enable = true
	config.Logs.Directory = "/tmp/gorobot-logs/"
	config.Logs.RecordEvents = true
	config.Logs.RecordMemoryUsage = true
	config.Logs.RecordStatistics = true

	config.Module.Interface = "localhost:3111"
	config.Module.AutoRunModules = true
	config.Module.AutoRun = []string {"rocket"}

	freenode := ConfigServer{
	Name: "freenode",
	Host: "irc.freenode.net",
	FloodControl: true,
	Nickname: "m1ch3l",
	Realname: "m1ch3l",
	Username: "m1ch3l",
	Password: "",
	Channels: make(map[string]*ConfigChannel),
	}

	sbrk := ConfigChannel{
	Name: "#sbrk",
	Password: "",
	Master: true,
	}
	freenode.Channels["sbrk"] = &sbrk

	config.Servers = make(map[string]*ConfigServer)
	config.Servers["freenode"] = &freenode

	return &config
}
