#!/usr/bin/lua

-- default template to write module in LUA

-- arg[1] = port
-- arg[2] = server
-- arg[3] = channel
-- arg[4] = user

-- sudo luarocks install luasocket
socket = require("socket")

if #arg < 4 then
   print("usage: " .. arg[0] .. " port server channel user")
   return false
end

port		= arg[1]
server		= arg[2]
channel		= arg[3]
user		= arg[4]
priority	= 1
params		= {}

for i = 4, #arg do
   params[i - 4] = arg[i]
   i = i + 1
end

-- create a connection to the administration port
sock = socket.connect("localhost", port)
if sock == nil then
   print("can't connect to localhost:" .. port)
   return false
end

-- send a message to the user on the channel where the command was invoked
function send_message(str)
      sock:send(string.format("%s %d PRIVMSG %s :%s -> %s\n", server, priority, channel, user, str))
end

-- send a message to the server where the command was invoked
function send_command(str)
   sock:send(string.format("%s %d %s\n", server, priority, str))
end

-- !roulette
-- randomly kick a player like russian roulette

math.randomseed(os.time())
if math.random(0, 7) == 0 then
   send_command(string.format("KICK %s %s :*PAN*", channel, user))
else
   send_message("*CLICK*")
end
