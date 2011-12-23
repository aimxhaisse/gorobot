<?php

$port		= $argv[1];
$server		= $argv[2];
$channel	= $argv[3];
$user		= $argv[4];
$priority	= 1;
$params		= array();

for ($i = 4; $i < sizeof($argv); ++$i)
    $params[$i - 4] = $argv[$i];

$sock = @fsockopen("localhost", $port);
if (!$sock)
    die(sprintf("can't connect to localhost %d\n", $port));

function	send_message($str)
{
    global	$sock, $priority, $channel, $user;

    fwrite($sock, sprintf("%s %d PRIVMSG %s :%s -> %s\n", $server, $priority, $channel, $user, $str));
}

function	send_private_message($str)
{
    global	$sock, $server, $priority, $channel, $user;

    fwrite($sock, sprintf("%s %d PRIVMSG %s :%s\n", $server, $priority, $user, $str));
}

function	send_command($str)
{
    global	$sock, $priority, $channel, $user;

    fwrite($sock, sprintf("%s %d %s\n", $server, $priority, $str));
}

function	send_raw($str)
{
    global	$sock, $priority, $channel, $user;

    fwrite($sock, $str);
}
