#!/usr/bin/env php
<?php
// command to kick a user

// usage string
if ($argc < 5)
    die("!kick server channel user message");

$port = $argv[1];
$serv = $argv[2];
$chan = $argv[3];
$user = $argv[4];

if ($argc < 9) {
    $message = sprintf("%s -> usage: !kick server channel user message", $user);
    $pre = "";
} else {
    $s = $argv[5];
    $c = $argv[6];
    $u = $argv[7];

    for ($i = 0; $i < 8; ++$i)
	unset($argv[$i]);

    $m = implode(" ", $argv);
    $message = sprintf("%s -> mission accomplished (or not)", $user);
    $pre = sprintf("%s 3 KICK %s %s :%s", $s, $c, $u, $m);
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
