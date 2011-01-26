#!/bin/sh

if [ $# -lt 8 ]
then
    echo "usage: !kick server chan user message"
    exit
fi

port=$1
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

echo "$server 3 KICK $chan $dest :$@" | nc -q 0 localhost $port > /dev/null
