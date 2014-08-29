package main

import (
	"strings"
	"fmt"
)

/*
 * Inspired by Golang's famous tutorial on Markov chains.
 */

type MarkovConfig struct {
	PrivateMessages MarkovConfigForEntry
	Targets map[string][]MarkovConfigForEntry
	PrefixLen		 int
	MaxWords	         int
}

type MarkovConfigForEntry struct {
	ChannelName              string
	AlwaysRespondToQuestions bool
	Verbosity		 int
	Record			 bool
}

type Prefix []string

func (p Prefix) String() string {
	return strings.Join(p, " ")
}

func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

type Chain struct {
	chain     map[string][]string
	prefixLen int
}

func (c *Chain) Insert(cfg *MarkovConfigForEntry, msg string, prefix_len int) {
	p := make(Prefix, c.prefixLen)
	for {
		var s string
		if _, err := fmt.Fscan(msg, &s); err != nil {
			break
		}
		key := p.String()
		c.chain[key] = append(c.chain[key], s)
		p.Shift(s)
	}
}

func (c *Chain) Answer(cfg *MarkovConfigForEntry, msg string, chac chan Action) {
}

func NewChain(prefixLen int) *Chain {
	return &Chain{make(map[string][]string), prefixLen}
}

func Markov(chac chan Action, chev chan Event, cfg MarkovConfig) {
	c := NewChain(cfg.PrefixLen)

	for {
		e := <-chev
		if e.Type == E_PRIVMSG {
			entries, ok := cfg.Targets[e.Server]
			if ok {
				for _, v := range(entries) {
					if (v.Record) {
						c.Insert(&v, e.Data)
					}
					c.Answer(&v, e.Data, chac)
				}
			}
		}
	}	
}
