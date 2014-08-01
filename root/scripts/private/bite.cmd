#!/usr/bin/env bash

po=$1
se=$2
ch=$3
us=$4

size=$(echo "$RANDOM % 20 + 3" | bc)
bite="B"
while [ $size -gt 0 ]
do
    bite="${bite}="
    size=$(echo $size - 1 | bc)
done
bite="${bite}D"

echo "$se 1 PRIVMSG $us :${us}> $bite" | nc -q 1 localhost $po
