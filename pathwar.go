package main

import (
	"fmt"
	"time"
	"net/http"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)

// Configuration of the pathwar module
type PathwarConfig struct {
	Targets map[string][]string	// List of servers/channels-users to broadcast
	EndPoint string			// https://user:token@pathwar-api-endpoint.tld/
}

type PathwarActivity struct {
	action string
	when string
}

func getActivities(client *http.Client, url string, last *time.Time) (*time.Time, []PathwarActivity) {
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("error: %v", err)
		return last, nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error: %v", err)
		return last, nil
	}

	var data map[string] interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("error: %v", err)
		return last, nil
	}

	results := make([]PathwarActivity, 0)

	if v, ok := data["_items"]; ok {
		switch vv := v.(type) {
		case []interface{}: 
			for _, entry := range(vv) {
				switch vvv := entry.(type) {
				case map[string] interface{}: 
					a := PathwarActivity{
						action: "",
						when: "",
					}
					if value, ok := vvv["action"]; ok {
						a.action = value.(string)
					}
					if value, ok := vvv["_created"]; ok {
						a.when = value.(string)
					}
					layout := "Mon, 02 Jan 2006 15:04:05 GMT"
					created_at, err := time.Parse(layout, a.when)
					if (err == nil && (last == nil || created_at.After(*last))) {
						results = append(results, a)
						last = &created_at
					}
				}
			}
		}
	}

	return last, results
}

// Broadcast listens for private messages and broadcasts them to a list of targets
func Pathwar(chac chan Action, config PathwarConfig) {
	a := Action{
		Type:     A_SAY,
		Priority: PRIORITY_LOW,
	}

	url := fmt.Sprintf("%s/activities?sort=-_updated", config.EndPoint)

	var last *time.Time
	var data []PathwarActivity

	for {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		last, data = getActivities(client, url, last)
		for _, entry := range(data) {
			what := fmt.Sprintf("NEWS M1CH3L %s (%s)\n", entry.action, entry.when)
			for server, targets := range config.Targets {
				a.Server = server
				a.Channel = ""
				a.User = ""
				for _, target := range targets {
					if strings.Index(target, "#") == 1 {
						a.Channel = target
					} else {
						a.User = target
					}
					a.Data = what
					chac <- a
				}
			}
		}
		time.Sleep(1 * 1e9)
	}
}
