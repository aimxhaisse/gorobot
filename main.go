package main

import (
	"gorobot"
)

func main() {
	michel := gorobot.NewGoRobot("config.json")
	michel.Run()
}
