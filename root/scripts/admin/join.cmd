#!/usr/bin/env php
<?php
// command to join a new channel

// usage string
if ($argc < 5)
    die("!join server channel");

$port = $argv[1];
$serv = $argv[2];
$chan = $argv[3];
$user = $argv[4];

if ($argc < 7) {
    $pre = "";
    $message = sprintf("%s -> usage: !join server channel", $user);
} else {
    $message = sprintf("%s -> channel joined", $user);
    $pre = sprintf("%s 3 JOIN %s\r\n", $argv[5], $argv[6]);
}

// send the cmd to m1ch3l
$cmd = $pre . sprintf("%s 1 PRIVMSG %s :%s\r\n", $serv, $chan, $message);
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
