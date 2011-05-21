package main

// module that only handles atom based feeds for now
// feel free to improve it

import (
	"botapi"
	"http"
	"xml"
	"fmt"
	"time"
	"strings"
	"log"
)

type RssFeed struct {
	Config ConfigFeed
	Name   string
	Items  map[string]string
}

type Feed struct {
	XMLName xml.Name "http://www.w3.org/2005/Atom feed"
	Item    []Item   "feed>entry"
}

type Item struct {
	Title string
	Link  string "id"
}

func InitFeeds(config *Config) *map[string]RssFeed {
	result := make(map[string]RssFeed)

	for name, feed := range config.Feeds {
		result[name] = RssFeed{
			Config: feed,
			Name:   name,
			Items:  make(map[string]string),
		}
	}

	return &result
}

func GetXmlFromUrl(url string) *Feed {
	r, _, err := http.Get(url)

	if err == nil {
		var feed Feed
		err := xml.Unmarshal(r.Body, &feed)
		if err != nil {
			log.Printf("%s\n", err)
		}
		r.Body.Close()
		return &feed
	}

	return nil
}

func PopulateFeed(feed *RssFeed) {
	fmt.Printf("Fetching...\n")
	feedxml := GetXmlFromUrl(feed.Config.Url)
	for _, item := range feedxml.Item {
		feed.Items[item.Link] = item.Title
		fmt.Printf("%s > %s\n", item.Link, item.Title)
	}
	fmt.Printf("OK\n")
}

func BroadcastNewItem(feed *ConfigFeed, chac chan botapi.Action, item *Item) {
	ac := botapi.Action{
		Data:     fmt.Sprintf("rss> %s [ %s ]", item.Title, item.Title),
		Priority: botapi.PRIORITY_LOW,
		Type:     botapi.A_SAY,
	}

	for srv, array := range feed.BroadCasts {
		ac.Server = srv
		for _, target := range array {
			if strings.Index(target, "#") == 1 {
				ac.Channel = target
			} else {
				ac.User = target
			}
			chac <- ac
		}
	}
}

func DrainFeed(feed RssFeed, chac chan botapi.Action) {
	for {
		feedxml := GetXmlFromUrl(feed.Config.Url)
		if feedxml != nil {
			for _, item := range feedxml.Item {
				if _, ok := feed.Items[item.Link]; ok == false {
					BroadcastNewItem(&feed.Config, chac, &item)
					feed.Items[item.Link] = item.Title
				}
			}
		}
		secs := 1e9 * feed.Config.Refresh
		time.Sleep(secs)
		log.Printf("Feed [%s] drained\n", feed.Name)
	}
}

func main() {
	config := NewConfig("./mod-rss.json")
	chac, chev := botapi.ImportFrom(config.RobotInterface, config.ModuleName)
	feeds := InitFeeds(config)

	for _, feed := range *feeds {
		PopulateFeed(&feed)
		go DrainFeed(feed, chac)
	}

	for {
		<-chev
	}
}
