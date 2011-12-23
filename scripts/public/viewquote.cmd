#!/usr/bin/env php
<?php

// tiny module to store quotes
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

if ($argc < 6) {
    $message = sprintf("%s -> usage: !viewquote id", $user);
} else {
    $id = $argv[5];
    $fp = fopen("quotes.db", "a+", 0666);
    if ($fp && flock($fp, LOCK_EX)) {
	$db = read_db($fp);
	var_dump($db);
	if (isset($db[$id])) {
	    $message = sprintf("%s -> quote %d [%s] (%s) %s",
			       $user, $id, $db[$id]['date'], $db[$id]['author'], $db[$id]['quote']);
	} else {
	    $message = sprintf("%s -> no quote for that id", $user);
	}
	flock($fp, LOCK_UN);
	fclose($fp);
    } else
	$message = sprintf("%s -> unable to view quote", $user);
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
