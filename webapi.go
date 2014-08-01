package main

import (
	"log"
	"net/http"
	"fmt"
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

// Config for the WebAPI
type WebAPIConfig struct {
	HTTPInterface    string			  // HTTP interface to use
	HTTPPort	 int			  // HTTP port to use
	HTTPServerName   string                   // Internal name of the server (should not conflict with server aliases)
}

type WebAPIHandler struct {
	Sessions	map[string][]Action
	Events		chan Event
	Actions		chan Action
}

func NewWebAPIHandler(ev chan Event, ac chan Action) *WebAPIHandler {
	return &WebAPIHandler{
		make(map[string][]Action),
		ev,
		ac,
	}
}

func (h *WebAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("received a request")
	fmt.Fprintf(w, "Welcome to the home page!")
}

func (h *WebAPIHandler) Loop() {
	for {
		select {
		case action, ok := <-h.Actions:
			if !ok {
				log.Printf("webapi action channel closed, bye bye")
				return
			}
			h.Sessions[action.User] = append(h.Sessions[action.User], action)
		}
	}
}

func WebAPI(cfg *WebAPIConfig, ev chan Event, ac chan Action) {
	listen_on := fmt.Sprintf("%s:%d", cfg.HTTPInterface, cfg.HTTPPort)
	log.Printf("WebAPI listens on %s", listen_on)
	if http.ListenAndServe(listen_on, NewWebAPIHandler(ev, ac)) != nil {
		log.Printf("webapi is not able to listen on %s, bye bye", listen_on)
		return
	}
}
