package main

import (
	"fmt"
	"time"
	"net/http"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
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

func getActivities(client *http.Client, url string, last time.Time) []PathwarActivity {
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil
	}

	var data map[string] interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil
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
					if err == nil && created_at.After(last) {
						results = append(results, a)
					}
				}
			}
		}
	}

	return results
}

// Broadcast listens for private messages and broadcasts them to a list of targets
func Pathwar(chac chan Action, config PathwarConfig) {
	url := fmt.Sprintf("%s/activities?sort=-_created", config.EndPoint)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	since := time.Now()

	for {
		client := &http.Client{Transport: tr}
		data := getActivities(client, url, since)
		fmt.Printf("I have %d activities\n", len(data))
		since = time.Now()
		time.Sleep(1 * 1e9)
	}
}
