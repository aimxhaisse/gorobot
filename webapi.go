package main

import (
	"time"
)

// Some context around this WebAPI.
//
// The current design is: we have IRC servers that poll events from
// the network which are parsed, and dispatched to a channel (Event),
// events are then processed by modules which may produce Actions over
// servers (for example: a kick).
//
// The WebAPI tries to mimmic an IRC server, it injects events in the
// processing chain and also consumes all actions producted by servers.
// A special server name is used for this WebAPI server, so that the WebAPI
// knows which actions are interesting (just like a regular IRC server,
// really).

type WebAPI struct {
}

func NewWebAPI(cfg *Config, chev chan Event) *WebAPI {
	r := &WebAPI{}

	go r.pollEvents(chev)

	return r
}

func (w *WebAPI) pollEvents(chev chan Event) {
	for {
		time.Sleep(1e9)
	}
}
