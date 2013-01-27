package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

// Events are built from the output of the IRC server, and are sent to modules
// Please keep this is in order of use, as some expression may overlap others
var re_server_notice = regexp.MustCompile("^:[^ ]+ NOTICE [^:]+ :(.*)")
var re_server_message = regexp.MustCompile("^:[^ ]+ ([0-9]+) [^:]+ :(.*)")
var re_server_ping = regexp.MustCompile("^PING :(.*)")
var re_event_join = regexp.MustCompile("^:([^!]+)![^ ]* JOIN :(.+)")
var re_event_part = regexp.MustCompile("^:([^!]+)![^ ]* PART ([^ ]+).*")
var re_event_privmsg = regexp.MustCompile("^:([^!]+)![^ ]* PRIVMSG ([^ ]+) :(.+)")
var re_event_kick = regexp.MustCompile("^:([^!]+)![^ ]* KICK ([^ ]+) ([^ ]+) :(.*)")
var re_event_quit = regexp.MustCompile("^:([^!]+)![^ ]* QUIT :(.*)")
var re_event_nick = regexp.MustCompile("^:([^!]+)![^ ]* NICK :(.*)")

func ExtractEvent(line string) *Event {
	if m := re_server_notice.FindStringSubmatch(line); len(m) == 2 {
		return newEventNOTICE(line, m[1], 0)
	}
	if m := re_server_message.FindStringSubmatch(line); len(m) == 3 {
		cmd_id, _ := strconv.Atoi(m[1])
		return newEventNOTICE(line, m[2], cmd_id)
	}
	if m := re_server_ping.FindStringSubmatch(line); len(m) == 2 {
		return newEventPING(line, m[1])
	}
	if m := re_event_join.FindStringSubmatch(line); len(m) == 3 {
		return newEventJOIN(line, m[1], m[2])
	}
	if m := re_event_part.FindStringSubmatch(line); len(m) == 3 {
		return newEventPART(line, m[1], m[2])
	}
	if m := re_event_privmsg.FindStringSubmatch(line); len(m) == 4 {
		return newEventPRIVMSG(line, m[1], m[2], m[3])
	}
	if m := re_event_kick.FindStringSubmatch(line); len(m) == 5 {
		return newEventKICK(line, m[1], m[2], m[3], m[4])
	}
	if m := re_event_quit.FindStringSubmatch(line); len(m) == 3 {
		return newEventQUIT(line, m[1], m[2])
	}
	if m := re_event_nick.FindStringSubmatch(line); len(m) == 3 {
		return newEventNICK(line, m[1], m[2])
	}
	log.Printf("Ignored message: %s", line)
	return nil
}

func newEventPING(line string, server string) *Event {
	event := new(Event)
	event.Raw = line
	event.Type = E_PING
	event.Data = server
	return event
}

func newEventNOTICE(line string, message string, cmd_id int) *Event {
	event := new(Event)
	event.Raw = line
	event.Type = E_NOTICE
	event.Data = message
	event.CmdId = cmd_id
	return event
}

func newEventJOIN(line string, user string, channel string) *Event {
	event := new(Event)
	event.Raw = line
	event.Type = E_JOIN
	event.Channel = channel
	event.Data = channel
	event.User = user
	return event
}

func newEventPART(line string, user string, channel string) *Event {
	event := new(Event)
	event.Raw = line
	event.Type = E_PART
	event.Channel = channel
	event.User = user
	return event
}

func newEventPRIVMSG(line string, user string, channel string, msg string) *Event {
	event := new(Event)
	event.Raw = line
	event.Type = E_PRIVMSG
	event.Data = msg
	if strings.Index(channel, "#") == 0 {
		event.Channel = channel
	}
	event.User = user
	return event
}

func newEventKICK(line string, user string, channel string, target string, msg string) *Event {
	event := new(Event)
	event.Raw = line
	event.Type = E_KICK
	event.Data = target
	event.Channel = channel
	event.User = user
	return event
}

func newEventQUIT(line string, user string, msg string) *Event {
	event := new(Event)
	event.Raw = line
	event.Type = E_QUIT
	event.Data = msg
	event.User = user
	return event
}

func newEventNICK(line string, user string, newuser string) *Event {
	event := new(Event)
	event.Raw = line
	event.Type = E_NICK
	event.Data = newuser
	event.User = user
	return event
}
