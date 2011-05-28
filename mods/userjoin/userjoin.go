package userjoin

import (
	"gorobot/api"
	"exec"
	"log"
)

func UserJoin(chac chan api.Action, chev chan api.Event, config Config) {
	for {
		e := <-chev
		// if the event is join, exec gurken
		if e.Type == api.E_JOIN {
			log.Printf("Penality for [%s]\n", e.User)
			path := "../../gurken"
			argv := []string{path, "userjoin", e.User}
			cmd, err := exec.Run(path, argv, []string{}, "", exec.Pipe, exec.Pipe, exec.Pipe)
			if err == nil {
				cmd.Wait(0)
			}
		}
	}
}
