package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
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
	HTTPInterface  string // HTTP interface to use
	HTTPPort       int    // HTTP port to use
	HTTPServerName string // Internal name of the server (should not conflict with server aliases)
}

type WebAPIHandler struct {
	Config       WebAPIConfig
	Sessions     map[string][]Action
	SessionsLock sync.Mutex
	Events       chan Event
	Actions      chan Action
}

type WebAPIRequest struct {
	Action string
	Login  string
	Data   string
}

type WebAPIResponse struct {
	ReturnCode string
	Messages   []string
}

func NewWebAPIHandler(cfg WebAPIConfig, ev chan Event, ac chan Action) *WebAPIHandler {
	return &WebAPIHandler{
		cfg,
		make(map[string][]Action),
		sync.Mutex{},
		ev,
		ac,
	}
}

func (h *WebAPIHandler) sendResponse(w http.ResponseWriter, rc string, messages []string) {
	rsp := WebAPIResponse{
		ReturnCode: rc,
		Messages:   messages,
	}
	w.Header().Add("content-type", "application/json")
	b, err := json.Marshal(&rsp)
	if err != nil {
		log.Printf("webapi: can't generate response: %v", err)
		return
	}
	w.Write(b)
}

func (h *WebAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	content := make([]byte, 1024)
	n, err := r.Body.Read(content)
	if err != nil || n == 0 {
		h.sendResponse(w, "RC_NOK", make([]string, 0))
		return
	}
	req := &WebAPIRequest{}
	err = json.Unmarshal(content[0:n], &req)
	if err != nil {
		h.sendResponse(w, "RC_NOK", make([]string, 0))
		return
	}

	if req.Action == "POLL" {
		h.SessionsLock.Lock()
		actions, ok := h.Sessions[req.Login]
		if ok {
			msgs := make([]string, 0)
			for _, action := range actions {
				msgs = append(msgs, action.Data)
			}
			h.Sessions[req.Login] = make([]Action, 0)
			h.SessionsLock.Unlock()
			h.sendResponse(w, "RC_OK", msgs)
		} else {
			h.SessionsLock.Unlock()
			h.sendResponse(w, "RC_OK", make([]string, 0))
		}
	} else if req.Action == "SAY" {
		go func(event chan Event, srv_name string, login string, data string) {
			event <- Event{false, srv_name, "", login, E_PRIVMSG, data, "", 42}
		}(h.Events, h.Config.HTTPServerName, req.Login, req.Data)
		h.sendResponse(w, "RC_OK", make([]string, 0))
	}
}

func (h *WebAPIHandler) Loop() {
	for {
		select {
		case action, ok := <-h.Actions:
			if !ok {
				log.Printf("webapi action channel closed, bye bye")
				return
			}
			h.SessionsLock.Lock()
			defer h.SessionsLock.Unlock()
			h.Sessions[action.User] = append(h.Sessions[action.User], action)
			h.SessionsLock.Unlock()
		}
	}
}

func WebAPI(cfg *WebAPIConfig, ev chan Event, ac chan Action) {
	listen_on := fmt.Sprintf("%s:%d", cfg.HTTPInterface, cfg.HTTPPort)
	log.Printf("WebAPI listens on %s", listen_on)
	handler := NewWebAPIHandler(*cfg, ev, ac)
	go handler.Loop()
	if http.ListenAndServe(listen_on, handler) != nil {
		log.Printf("webapi is not able to listen on %s, bye bye", listen_on)
		return
	}
}
