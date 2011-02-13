#!/usr/bin/env bash

po=$1
se=$2
ch=$3
us=$4

if [ $# -lt 8 ]
then
    echo "$se 3 PRIVMSG $ch :$us: !kick server chan user message" | nc -q 0 localhost $po > /dev/null
    exit
fi

server=$5
chan=$6
dest=$7

shift
shift
shift
shift
shift
shift
shift

echo "$server 3 KICK $chan $dest :$@" | nc -q 0 localhost $po > /dev/null
