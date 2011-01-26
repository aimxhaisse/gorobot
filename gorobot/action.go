package gorobot

import(
	"regexp"
	"strings"
)

// Actions are a way for modules to communicate with the bot, basically
// they are just sent on a channel and interpreted.

var re_cmd_join	        = regexp.MustCompile("^JOIN ([^ ]+).*")
var re_cmd_kick		= regexp.MustCompile("^KICK ([^ ]+) ([^ ]+) :(.*)")
var re_cmd_kick_short	= regexp.MustCompile("^KICK ([^ ]+) ([^ ]+).*")
var re_cmd_part		= regexp.MustCompile("^PART ([^ ]+) :(.*)")
var re_cmd_part_short	= regexp.MustCompile("^PART ([^ ]+).*")
var re_cmd_privmsg_chan	= regexp.MustCompile("^PRIVMSG ([^ ]+) :(.*)")

// Constructs an Action of any type from a Raw Action, this extra step is
// made to ensure that the bot is still aware of what's hapening even with
// raw actions (ie: a raw action "QUIT" has to remove the server from the bot)
func ExtractAction(raw_action *Action) (*Action) {
	if m := re_cmd_kick.FindStringSubmatch(raw_action.Data); len(m) == 4 {
		return ActionKICK(&raw_action.Server, &m[1], &m[2], &m[3])
	}
	if m := re_cmd_kick_short.FindStringSubmatch(raw_action.Data); len(m) == 3 {
		return ActionKICK(&raw_action.Server, &m[1], &m[2], nil)
	}
	if m := re_cmd_join.FindStringSubmatch(raw_action.Data); len(m) == 2 {
		return ActionJOIN(&raw_action.Server, &m[1])
	}
	if m := re_cmd_part.FindStringSubmatch(raw_action.Data); len(m) == 3 {
		return ActionPART(&raw_action.Server, &m[1], &m[2])
	}
	if m := re_cmd_part_short.FindStringSubmatch(raw_action.Data); len(m) == 2 {
		return ActionPART(&raw_action.Server, &m[1], nil)
	}
	if m := re_cmd_privmsg_chan.FindStringSubmatch(raw_action.Data); len(m) == 3 {
		return ActionPRIVMSG(&raw_action.Server, &m[1], &m[2])
	}
	return nil
}

func ActionKICK(srv *string, channel *string, user *string, msg *string) (*Action) {
	result := new(Action)
	result.Server = *srv
	result.Channel = *channel
	result.User = *user
	if msg != nil {
		result.Data = *msg
	}
	result.Type = A_KICK
	return result
}

func ActionJOIN(srv *string, channel *string) (*Action) {
	result := new(Action)
	result.Server = *srv
	result.Channel = *channel
	result.Type = A_JOIN
	return result
}

func ActionPART(srv *string, channel *string, msg *string) (*Action) {
	result := new(Action)
	result.Server = *srv
	result.Channel = *channel
	if msg != nil {
		result.Data = *msg
	}
	result.Type = A_PART
	return result
}

func ActionPRIVMSG(srv *string, channel *string, msg *string) (*Action) {
	result := new(Action)
	result.Server = *srv
	if strings.Index(*channel, "#") == 0 {
		result.Channel = *channel
	} else {
		result.User = *channel
	}
	result.Data = *msg
	result.Type = A_SAY
	return result
}
