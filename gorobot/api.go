package gorobot

import (
	"netchan"
	"log"
	"time"
)

type EventType int
const (
	E_PRIVMSG = iota // A PRIVMSG sentin private or on a channel
	E_NOTICE // Same as PRIVMSG but for notices
	E_JOIN // A user joined (origin is the channel, data is the user)
	E_PART // A user leaved a chan (same thing)
	E_QUIT // A user quit (origin is the user)
	E_NICK // A user changed nick (origin is old nick, data is the new nick)
	E_PING // A ping has been sent by the server
)

type ActionType int
const (
	A_SAY = iota // Say something
	A_KICK
	A_JOIN
	A_PART
	A_OP
	A_RAW
	A_SENDNOTICE
	A_NEWMODULE // This is sent by a module after it connects for
		    // the first time to m1ch3l. This is handled
		    // internally, modules should not care about this
		    // value.
)

type Event struct {
	AdminCmd bool // Is this event admin-issued ?
	Server string // The server on which the event occured
	Channel string // The #channel on which the event occured
	User string // Nickname of the user who triggered the event
	Type EventType // Type of the event
	Data string // Additional data
	Raw string // Raw command
	CmdId int // Id of the command
}

const (
	PRIORITY_LOW = 1 // the action has a low priority
	PRIORITY_MEDIUM = 2
	PRIORITY_HIGH = 3
)

type Action struct {
	Server string // What server to operate on
	Channel string // What channel to operate on
	User string // Who is concerned
	Data string // Additional data
	Priority int // priority of the message (if type == SAY)
	Type ActionType // What to do
	Raw string // If Type = RAW, send this directly over the network
}

func ImportFrom(hostname string, moduleUUID string) (chan Action, chan Event) {
	imp, err := netchan.NewImporter("tcp", hostname)
	if err != nil {
		log.Exit(err)
	}
	chac := make(chan Action)
	err = imp.Import("actions", chac, netchan.Send, -1)
	if err != nil {
		log.Exit(err)
	}

	id := Action{Type: A_NEWMODULE, Data: moduleUUID}
	chac <- id

	// UGLY UGLY UGLY
	time.Sleep(100000000)

	chev := make(chan Event)
	err = imp.Import("events-" + moduleUUID, chev, netchan.Recv, -1)
	if err != nil {
		log.Exit(err)
	}
	return chac, chev
}

func InitExport(bindAddr string) (*netchan.Exporter) {
        exp, err := netchan.NewExporter("tcp", bindAddr)
	if err != nil {
		log.Exit(err)
	}
	return exp
}

func ExportActions(exp *netchan.Exporter) (chan Action) {
	chac := make(chan Action)
	err := exp.Export("actions", chac, netchan.Recv)
	if err != nil {
		log.Exit(err)
	}

	go func(){
		for {
			exp.Drain(-1)
			time.Sleep(1000000)
		}
	}()
	return chac
}

func ExportEvents(exp *netchan.Exporter, moduleUUID string) (chan Event) {
	chev := make(chan Event)
	err := exp.Export("events-" + moduleUUID, chev, netchan.Send)
	if err != nil {
		exp.Hangup("events-" + moduleUUID)
		err := exp.Export("events-" + moduleUUID, chev, netchan.Send)
		if err != nil {
			log.Exit(err)
		}
	}
	go func() {
		for {
			exp.Drain(-1)
			time.Sleep(1000000)
		}
	}()
	return chev
}
