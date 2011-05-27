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

> make install && cd bin && ed gorobot.json && ed rocket.json && ./gorobot

## FAQ

### Can I add a new command?

Yes.

### Can I create a new module?

Yes.

### Can I restart modules without restarting the bot?

Yes.

### Is there a documentation with real answers?

No.
