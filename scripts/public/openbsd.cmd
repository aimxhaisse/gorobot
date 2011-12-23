#!/usr/bin/env lua

package.path = package.path .. ";scripts/?.lua"
require("helper")

math.randomseed(os.time())
val = math.random(1, 30)

send_message("sur openbsd on envoyer des mail anonyme")
