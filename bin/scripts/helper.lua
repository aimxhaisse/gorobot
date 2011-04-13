-- sudo luarocks install luasocket
socket = require("socket")

if #arg < 4 then
   error(string.format("usage: %s port server channel user", arg[0]))
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
   error(string.format("can't connect to localhost: %d", port))
end

-- send a message to the user on the channel where the command was invoked
function send_message(str)
      sock:send(string.format("%s %d PRIVMSG %s :%s -> %s\n", server, priority, channel, user, str))
end

-- send a message to the server where the command was invoked
function send_command(str)
   sock:send(string.format("%s %d %s\n", server, priority, str))
end
