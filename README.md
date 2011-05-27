GoRobot
===

## an IRC bot written in Go

## Features:

  * Multiple servers
  * Multiple channels
  * Conversations
  * Administration via a specific set of IRC channels
  * Go modules, which can connect to the bot through netchans (one can dynamically add new modules through network)
  * Simple API to write GO modules
  * JSON Configuration
  * Flood control
  * Module to handle shell scripts, through a tiny API
  * Module to follow RSS feeds
  * Module to follow MPD stream
  * Statistics (activity on a channel, number of people, ...)

## What are these folders?

  * bin/ stores binaries once compiled (bot and mods), shell scripts for commands, config files
  * api/ stores sources of the go API (used by mods to dialog with gorobot)
  * bot/ stores sources for the IRC robot
  * mods/ stores sources for modules (each module is a package, to run a module add it to the rocket)
  * rocket/ stores sources for a launcher of modules

## Installation

```sh
make install
cd bin
ed gorobot.json
ed rocket.json
./gorobot
```
## Commands

### How it works

Commands can be added in folders bin/scripts/{admin,public,private}.

Private commands are executed when talking in private with the bot.
Public commands are executed on all channels.
Admin commands are executed 

### Available commands

Private: !spoon
Public: !chat !non !pokemon !roulette !viewquote !ninja !fax !pwet !boby !matrix !oui !template !statquote ...
Admin: !addquote !join !kick !part

(mostly lame and useless commands)

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

po=$1
se=$2
ch=$3
us=$4

echo "$se 1 PRIVMSG $us :th3r3 1s n0 sp0on..." | nc localhost $po
```

Once the command is created, don't forget to chmod it (+x).

## FAQ

### Can I add a new command?

Yes.

### Can I create a new module?

Yes.

### Can I restart modules without restarting the bot?

Yes.

### Is there a documentation with real answers?

No.
