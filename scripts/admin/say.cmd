#!/usr/bin/env bash

po=$1
se=$2
ch=$3
us=$4

if [ $# -lt 7 ]
then
    echo "usage: !say server receiver message" | nc -q 0 localhost $po > /dev/null
    exit
fi

server=$5
dest=$6

shift
shift
shift
shift
shift
shift

echo "$server 1 PRIVMSG $dest :$@" | nc -q 0 localhost $po > /dev/null
