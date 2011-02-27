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
    return unserialize($buffer);
}

function write_db($fp, $newdb)
{
    $string = serialize($newdb);
    for ($written = 0; $written < strlen($string); $written += $fwrite) {
        $fwrite = fwrite($fp, substr($string, $written));
        if ($fwrite === false) {
	    break;
        }
    }
}

// try to store a new quote in the serialized file
if (!isset($argv[5])) {
    $message = sprintf("%s -> usage: !quote_add quote", $user);
} else {
    $quote = $argv[5];
    $fp = fopen("quotes.db", "a+", 0666);
    if ($fp && flock($fp, LOCK_EX)) {
	$db = read_db($fp);
	if (!$db)
	    $db = array();
	$db[] = array(
	    'server' => $serv,
	    'chan' => $chan,
	    'added_by' => $user,
	    'quote' => $quote
	    );
	ftruncate($fp, 0);
	write_db($fp, $db);
	flock($fp, LOCK_UN);
	fclose($fp);
	$message = sprintf("%s -> quote added", $user);
    } else
	$message = sprintf("%s -> unable to add quote", $user);
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
