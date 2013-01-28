#!/usr/bin/env lua

-- !roulette
-- randomly kick a player like russian roulette

package.path = package.path .. ";scripts/?.lua"
require("helper")

math.randomseed(os.time())
if math.random(0, 7) == 0 then
   send_command(string.format("KICK %s %s :*PAN*", channel, user))
else
   send_message("*CLICK*")
end
