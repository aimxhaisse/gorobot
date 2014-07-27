GoRobot
===

Yet Another IRC robot.

## Features:

  * Multiple servers, multiple channels, conversations, flood control
  * Administration via a specific set of IRC channels
  * JSON configuration
  * Module to handle shell scripts, through a tiny API

Docker
------

    # Build
    docker build -t aimxhaisse/gorobot .

    # Run in foreground
    docker run -i -t -rm aimxhaisse/gorobot

    # Run in background
    docker run -d aimxhaisse/gorobot

    # Mounts scripts directory for dev
    docker run -i -t -rm \
    	   -v $(pwd)/root/ /home/gorobot/gorobot/root/ \
    	   aimxhaisse/gorobot
    	   
Extending with Docker
---------------------

    FROM aimxhaisse/gorobot
    ADD . ./root
    ...

## Commands

### How it works

Commands can be added in folders scripts/{admin,public,private}.

  * Private commands are executed when talking in private with the bot.
  * Public commands are executed on all channels.
  * Admin commands are executed on master channels (see grobot.json).

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

Server is the server where the command has to be executed, priority is
a number (1, 2 or 3) indicating the priority of the command. This
priority is meaningful on servers having flood control (you may want
to kick someone before printing 42 lines).

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
