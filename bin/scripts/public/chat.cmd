#!/usr/bin/env lua

-- !roulette
-- randomly kick a player like russian roulette

package.path = package.path .. ";scripts/?.lua"
require("helper")

math.randomseed(os.time())
val = math.random(1, 30)

send_message("les chats c'est des connards")
