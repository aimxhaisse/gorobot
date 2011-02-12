package main

import (
	"gorobot"
)

func main() {
	bot := gorobot.NewGoRobot("config.json")
	bot.Run()
}
