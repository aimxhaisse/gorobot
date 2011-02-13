package main

import (
	"json"
	"io/ioutil"
	"log"
)

type Users struct {
	Accounts	map[string] string
	Authenticated	map[string] bool
}

func NewUsers() *Users {
	u := Users{
		Accounts: make(map[string] string),
		Authenticated: make(map[string] bool),
	}
	return &u
}

func (users *Users) IsAuthenticated(username string) bool {
	auth, ok := users.Authenticated[username]
	return ok && auth
}

func (users *Users) Login(username string, login string, pass string) {
	log.Printf("%s-%s", login, pass)
	if _, ok := users.Accounts[login]; ok == true {
		if users.Accounts[login] == pass {
			users.Authenticated[username] = true
			log.Printf("%s is now authenticated (%s)", login, username)
		} else {
			log.Printf("authentication failed for %s (%s)", login, username)
		}
	}
}

func (users *Users) Rename(oldname string, newname string) {
	if _, ok := users.Authenticated[oldname]; ok == true {
		if _, ok := users.Authenticated[newname]; ok == false {
			users.Authenticated[newname] = users.Authenticated[oldname]
			users.Logout(oldname)
		}
	}
}

func (users *Users) Logout(username string) {
	if _, ok := users.Authenticated[username]; ok == true {
		users.Authenticated[username] = false, false
	}
}

func (users *Users) Refresh(path string) {
	file, e := ioutil.ReadFile(path)
	if e != nil {
		log.Printf("users_db error: %v", e)
	}
	e = json.Unmarshal(file, &users.Accounts)
	if e != nil {
		log.Printf("users_db error: %v", e)
	}
	log.Printf("users_db refreshed")
}
