#!/usr/bin/env php
<?php

// !pokemon #number
// send a pokemon in private

require_once(sprintf('%s/%s', __DIR__, '../helper.php'));

if ($argc != 6)
{
    send_message("usage: !pokemon number");
    exit;
}

$pokemon = sprintf("%03d", $argv[5]);
$content = file_get_contents("http://www.angelfire.com/mn/Maija/pokemon/$pokemon.txt");

if (strlen($content) > 0)
{
    $lines = explode("\n", $content);
    $content = "";
    foreach ($lines as $line)
	$content .= sprintf("%s %d PRIVMSG %s :%s\n", $server, $priority, $user, $line);
    send_raw($content);
}
else
    send_private_message("missing no.");
