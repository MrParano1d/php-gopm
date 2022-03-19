<?php

declare(strict_types=1);

use GOPM\Server;
use GuzzleHttp\Psr7\Utils;

require_once __DIR__ . "./vendor/autoload.php";

$server = new Server();

$server->onRequest(function(\GuzzleHttp\Psr7\Request $request, \GuzzleHttp\Psr7\Response $response) {

  $response = $response->withHeader("Content-Type", "application/json");
  $response = $response->withBody(Utils::streamFor(json_encode(["hello" => "world 🌍"])));

  return $response;
});

$server->listen();
