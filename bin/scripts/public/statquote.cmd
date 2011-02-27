#!/usr/bin/env php
<?php

// tiny module to dump stats about quotes
// ugly storage, ugly code, that's because I'm ugly

if ($argc < 5)
    die("don't run me like this\n");

$port = $argv[1];
$serv = $argv[2];
$chan = $argv[3];
$user = $argv[4];

function read_db($fp)
{
    $result = '';
    while (!feof($fp)) {
	$result .= fread($fp, 8192);
    }
    return unserialize($result);
}

if ($argc >= 6) {
    $message = sprintf("%s -> usage: !statquote", $user);
} else {
    $fp = fopen("quotes.db", "a+", 0666);
    if ($fp && flock($fp, LOCK_EX)) {
	$db = read_db($fp);
	$message = sprintf("%s -> there is/are %d quote(s)", $user, count($db));
	flock($fp, LOCK_UN);
	fclose($fp);
    } else
	$message = sprintf("%s -> unable to view statistics", $user);
}

// send the cmd to m1ch3l
$cmd = sprintf("%s 1 PRIVMSG %s :%s\r\n", $serv, $chan, $message);
$sock = fsockopen("localhost", $port);
if ($sock) {
    $i = 0;
    while ($i < strlen($cmd)) {
        $n = fwrite($sock, $cmd);
        if ($n == false)
	    break;
        $i += $n;
    }
    fclose($sock);
}
