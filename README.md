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
  * No excess flood, even upon high activity
  * Module to handle shell scripts, through a tiny API
  * Statistics (activity on a channel, number of people, ...)

## Todo:

  * Module to follow RSS feeds
  * Module to offer a graphical interface through network (in progress)
  * A prompt to perform administrative tasks?
  * Automatic reconnection on timeout, kicks, ...
  * Add new features to the GO API (configuration of modules, ...)
  * Ability to autoload GO modules
