#!/usr/bin/env lua

-- !roulette
-- randomly kick a player like russian roulette

package.path = package.path .. ";scripts/?.lua"
require("helper")

math.randomseed(os.time())
val = math.random(1, 30)

bite = ''
for i = 0, val do
   bite = bite .. '='
end

send_message(string.format("8%sD", bite))
