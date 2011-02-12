package main

import (
	"botapi"
	"http"
	"xml"
	"fmt"
	"time"
	"strings"
)

type RssFeed struct {
	Config		ConfigFeed
	Name		string
	Items		map[string] string
}

type Feed struct {
	Item		[]Item "channel>item"
}

type Item struct {
	Title		string
	Link		string
}

func InitFeeds(config *Config) (*map[string] RssFeed) {
	result := make(map[string] RssFeed)

	for name, feed := range config.Feeds {
		result[name] = RssFeed{
			Config: feed,
			Name: name,
			Items: make(map[string] string),
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
			fmt.Printf("%s\n", err)
		}
		r.Body.Close();
		return &feed
	}

	return nil
}

func PopulateFeed(feed *RssFeed) {
	feedxml := GetXmlFromUrl(feed.Config.Url)
	for _, item := range feedxml.Item {
		feed.Items[item.Link] = item.Title
	}
}

func BroadcastNewItem(feed *ConfigFeed, chac chan botapi.Action, item *Item) {
	ac := botapi.Action {
		Data: fmt.Sprintf("rss> %s [ %s ]", item.Title, item.Link),
		Priority: botapi.PRIORITY_LOW,
		Type: botapi.A_SAY,
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
		time.Sleep(1e7 * feed.Config.Refresh)
		fmt.Printf("Feed [%s] drained\n", feed.Name)
	}
}

func main() {
	config := NewConfig("config.json")
	chac, chev := botapi.ImportFrom(config.RobotInterface, config.ModuleName)
	feeds := InitFeeds(config)

	for _, feed := range *feeds {
		PopulateFeed(&feed)
		go DrainFeed(feed, chac)
	}

	for {
		<- chev
	}
}
