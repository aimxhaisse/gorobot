package gorobot

import (
	"botapi"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// Events are built from the output of the IRC server, and are sent to modules
// Please keep this is in order of use, as some expression may overlap others
var re_server_notice       = regexp.MustCompile("^:[^ ]+ NOTICE [^:]+ :(.*)")
var re_server_message	   = regexp.MustCompile("^:[^ ]+ ([0-9]+) [^:]+ :(.*)")
var re_server_ping	   = regexp.MustCompile("^PING :(.*)")
var re_event_join	   = regexp.MustCompile("^:([^!]+)![^ ]* JOIN :(.+)")
var re_event_part	   = regexp.MustCompile("^:([^!]+)![^ ]* PART ([^ ]+).*")
var re_event_privmsg       = regexp.MustCompile("^:([^!]+)![^ ]* PRIVMSG ([^ ]+) :(.+)")
var re_event_kick	   = regexp.MustCompile("^:([^!]+)![^ ]* KICK ([^ ]+) ([^ ]+) :(.*)" )
var re_event_quit	   = regexp.MustCompile("^:([^!]+)![^ ]* QUIT :(.*)")

func ExtractEvent(line string) (*botapi.Event) {
	if m := re_server_notice.FindStringSubmatch(line); len(m) == 2 {
		return EventNOTICE(line, m[1], 0)
	}
	if m := re_server_message.FindStringSubmatch(line); len(m) == 3 {
		cmd_id, _ := strconv.Atoi(m[1])
		return EventNOTICE(line, m[2], cmd_id)
	}
	if m := re_server_ping.FindStringSubmatch(line); len(m) == 2 {
		return EventPING(line, m[1])
	}
	if m := re_event_join.FindStringSubmatch(line); len(m) == 3 {
		return EventJOIN(line, m[1], m[2])
	}
	if m := re_event_part.FindStringSubmatch(line); len(m) == 3 {
		return EventPART(line, m[1], m[2])
	}
	if m := re_event_privmsg.FindStringSubmatch(line); len(m) == 4 {
		return EventPRIVMSG(line, m[1], m[2], m[3])
	}
	if m := re_event_kick.FindStringSubmatch(line); len(m) == 5 {
		return EventKICK(line, m[1], m[2], m[3], m[4])
	}
	if m := re_event_quit.FindStringSubmatch(line); len(m) == 3 {
		return EventQUIT(line, m[1], m[2])
	}
	log.Printf("Ignored message: %s\n", line)
	return nil
}

func EventPING(line string, server string) (*botapi.Event) {
	event := new(botapi.Event)
	event.Raw = line
	event.Type = botapi.E_PING
	event.Data = server
	return event
}

func EventNOTICE(line string, message string, cmd_id int) (*botapi.Event) {
	event := new(botapi.Event)
	event.Raw = line
	event.Type = botapi.E_NOTICE
	event.Data = message
	event.CmdId = cmd_id
	return event
}

func EventJOIN(line string, user string, channel string) (*botapi.Event) {
	event := new(botapi.Event)
	event.Raw = line
	event.Type = botapi.E_JOIN
	event.Channel = channel
	event.Data = channel
	event.User = user
	return event
}

func EventPART(line string, user string, channel string) (*botapi.Event) {
	event := new(botapi.Event)
	event.Raw = line
	event.Type = botapi.E_PART
	event.Channel = channel
	event.User = user
	return event
}

func EventPRIVMSG(line string, user string, channel string, msg string) (*botapi.Event) {
	event := new(botapi.Event)
	event.Raw = line
	event.Type = botapi.E_PRIVMSG
	event.Data = msg
	if strings.Index(channel, "#") == 0 {
		event.Channel = channel
	}
	event.User = user
	return event
}

func EventKICK(line string, user string, channel string, target string, msg string) (*botapi.Event) {
	event := new(botapi.Event)
	event.Raw = line
	event.Type = botapi.E_KICK
	event.Data = target
	event.Channel = channel
	event.User = user
	return event
}

func EventQUIT(line string, user string, msg string) (*botapi.Event) {
	event := new(botapi.Event)
	event.Raw = line
	event.Type = botapi.E_QUIT
	event.Data = msg
	event.User = user
	return event
}
