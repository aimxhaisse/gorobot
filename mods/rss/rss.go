// package rss implements a module that broadcast atom based feeds for now
// feel free to improve it
package rss

import (
	"gorobot/api"
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

func initFeeds(config Config) *map[string]RssFeed {
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

func getXmlFromUrl(url string) *Feed {
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

func populateFeed(feed *RssFeed) {
	fmt.Printf("Fetching...\n")
	feedxml := getXmlFromUrl(feed.Config.Url)
	for _, item := range feedxml.Item {
		feed.Items[item.Link] = item.Title
		fmt.Printf("%s > %s\n", item.Link, item.Title)
	}
	fmt.Printf("OK\n")
}

func broadcastNewItem(feed *ConfigFeed, chac chan api.Action, item *Item) {
	ac := api.Action{
		Data:     fmt.Sprintf("rss> %s [ %s ]", item.Title, item.Title),
		Priority: api.PRIORITY_LOW,
		Type:     api.A_SAY,
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

func drainFeed(feed RssFeed, chac chan api.Action) {
	for {
		feedxml := getXmlFromUrl(feed.Config.Url)
		if feedxml != nil {
			for _, item := range feedxml.Item {
				if _, ok := feed.Items[item.Link]; ok == false {
					broadcastNewItem(&feed.Config, chac, &item)
					feed.Items[item.Link] = item.Title
				}
			}
		}
		secs := 1e9 * feed.Config.Refresh
		time.Sleep(secs)
		log.Printf("Feed [%s] drained\n", feed.Name)
	}
}

func Rss(chac chan api.Action, chev chan api.Event, config Config) {
	feeds := initFeeds(config)

	for _, feed := range *feeds {
		populateFeed(&feed)
		go drainFeed(feed, chac)
	}

	for {
		<-chev
	}
}
