#!/usr/bin/env bash

po=$1
se=$2
ch=$3
us=$4

if [ $# -lt 6 ]
then
    echo "$se 3 PRIVMSG $ch :usage: !join server chan [message]" | nc -q 0 localhost $po > /dev/null
    exit
fi

server=$5
chan=$6
msg=$7

shift
shift
shift
shift
shift
shift

if [ $# -eq 5 ]
then
    echo "$server 3 JOIN $chan" | nc -q 0 localhost $po > /dev/null
else
    echo "$server 3 JOIN $chan" | nc -q 0 localhost $po > /dev/null
    sleep 2 # ugly
    echo "$server 3 PRIVMSG $chan :$@" | nc -q 0 localhost $po > /dev/null
fi
