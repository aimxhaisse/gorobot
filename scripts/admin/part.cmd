#!/bin/sh

if [ $# -lt 6 ]
then
    echo "usage: !part server chan [message]"
    exit
fi

port=$1
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
    echo "$server 3 PART $chan" | nc -q 0 localhost $port > /dev/null
else
    echo "$server 3 PART $chan :$@" | nc -q 0 localhost $port > /dev/null
fi
