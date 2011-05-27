package gorobot

import (
	"gorobot"
)

func main() {
	bot := gorobot.NewGoRobot("gorobot.json")
	bot.Run()
}
