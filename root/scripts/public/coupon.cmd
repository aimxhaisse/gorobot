#!/usr/bin/env bash

po=$1
se=$2
ch=$3
us=$4

hash=$(echo "$RANDOM" | md5sum | cut -d' ' -f1)

echo "$se 1 PRIVMSG $ch :coupon: $hash" | nc -q 1 localhost $po
