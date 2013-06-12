#!/usr/bin/lua

-- default template to write module in LUA

-- arg[1] = port
-- arg[2] = server
-- arg[3] = channel
-- arg[4] = user

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

sock = socket.connect("localhost", port)
if sock == nil then
   print("can't connect to localhost:" .. port)
   return false
end

function send_message(str)
      sock:send(string.format("%s %d PRIVMSG %s :%s -> %s\n", server, priority, channel, user, str))
end

function send_command(str)
   sock:send(string.format("%s %d %s\n", server, priority, str))
end
