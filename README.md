GoRobot
===

## an IRC bot written in Go

GoRobot is an IRC robot aiming at being highly modular without any
need to disconnect.  To meet that purpose, it is composed of two
binaries :

  * grobot, the core server which maintains connections
  * grocket, a module composed of submodules that connects to grobot

Communication is made with the use of netchan, thus they don't need to
be on the same computer. There can be several grockets for one grobot.

## Features:

  * Multiple servers, multiple channels, conversations, flood control
  * Administration via a specific set of IRC channels
  * Go modules, which can connect to the bot through netchans (one can dynamically add new modules through network)
  * Simple API to write GO modules
  * JSON configuration
  * Module to handle shell scripts, through a tiny API
  * Module to follow RSS feeds
  * Module to follow MPD stream
  * Statistics (activity on a channel, memory usage, ...)

## What are these folders?

  * bin/ stores binaries once compiled (grobot and grocket)
  * api/ stores sources of the go API (used by mods to dialog with gorobot)
  * grobot/ stores sources for the IRC robot
  * mods/ stores sources for modules (each module is a package, to run a module add it to the grocket)
  * grocket/ stores sources for a launcher of modules
  * grobot.json is the default configuration file for grobot
  * grocket.json is the default configuration file for grobot

## Installation

```sh
git clone git://github.com/aimxhaisse/gorobot.git gorobot && cd gorobot
./build install

ed grobot.json
ed grocket.json

# in one term
grobot

# in another term
grocket
```
## Commands

### How it works

Commands are implemented in the mods/scripts module automatically launched by grocket.

Commands can be added in folders scripts/{admin,public,private}.

  * Private commands are executed when talking in private with the bot.
  * Public commands are executed on all channels.
  * Admin commands are executed on master channels (see gorobot.json).

### Available commands

Private: !spoon

Public: !chat !non !pokemon !roulette !viewquote !ninja !fax !pwet !boby !matrix !oui !template !statquote ...

Admin: !addquote !join !kick !part

### How to add new commands

You can add new commands in whatever language you want. Current ones are
in PHP or Lua (with some helpers to do the dirty job). Commands are executed
like this:

```sh
./bin/scripts/xxx/yyy.cmd <port> <server> <channel> <user> <param1> <param2> <...>
```

Example, "UserA" invokes "!hejsan 42" on the channel #toto42 of freenode:

```sh
./bin/scripts/xxx/yyy.cmd 2345 freenode #toto42 UserA 42
```

The port is a local port opened by the module "scripts", it accepts raw IRC commands in the following way:

```sh
<server> <priority> RAW_COMMAND
```

Server is the server where the command has to be executed, priority is a number (1, 2 or 3)
indicating the priority of the command. This priority is meaningful on servers having
flood control (you may want to kick someone before printing 42 lines).

Example of a bash command:

```sh
#!/usr/bin/env bash

port=$1
serv=$2
chan=$3
user=$4

echo "$serv 1 PRIVMSG $user :th3r3 1s n0 sp0on..." | nc localhost $po
```

Once the command is created, don't forget to chmod it (+x).
