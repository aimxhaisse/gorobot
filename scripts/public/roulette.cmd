#!/usr/bin/env bash

po=$1
se=$2
ch=$3
us=$4

if [ ! -f "public/roulette.txt" ]; then
    let "rand=$RANDOM % 6"
    echo $rand > public/roulette.txt
fi

count="`cat public/roulette.txt`"

if [ $count -eq 0 ]; then
    echo "$se 1 KICK $ch $us :*PAN*" | nc -q 0 localhost $po > /dev/null
    rm public/roulette.txt
else
    echo "$se 1 PRIVMSG $ch :*CLICK*" | nc -q 0 localhost $po > /dev/null
    count=$(($count - 1))
    echo $count > public/roulette.txt
fi
