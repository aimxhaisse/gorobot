#!/usr/bin/env lua

package.path = package.path .. ";scripts/?.lua"
require("helper")

math.randomseed(os.time())
val = math.random(1, 30)

send_message("les chats c'est des connards")
