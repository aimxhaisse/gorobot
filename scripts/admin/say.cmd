#!/bin/sh

if [ $# -lt 7 ]
then
    echo "usage: !say server receiver message"
    exit
fi

port=$1
server=$5
dest=$6

shift
shift
shift
shift
shift
shift

echo "$server 1 PRIVMSG $dest :$@" | nc -q 0 localhost $port > /dev/null
