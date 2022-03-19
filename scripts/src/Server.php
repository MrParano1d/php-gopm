<?php

declare(strict_types=1);

namespace GOPM;

class Server {

  public const STREAM_SOCKET_DEFAULT = "tcp://localhost:13337";

  private string $streamSocket = Server::STREAM_SOCKET_DEFAULT;

  private $httpHandler = null;

  public function __construct()
  {
    
  }

  public function onRequest(callable $handler) {
    $this->httpHandler = $handler;
  }

  public function listen() {
    $errorCode = 0;
    $errorString = "";
    $sock = stream_socket_client($this->streamSocket, $errorCode, $errorString, 10, STREAM_CLIENT_CONNECT | STREAM_CLIENT_PERSISTENT);

    if (!$sock) {
      echo "$errorString ($errorCode)\n";
      exit(1);
    }

    while ($reqStr = fread($sock, 1024)) {
      if ($reqStr === "ping") {
        continue;
      }
      $request = \GuzzleHttp\Psr7\Message::parseRequest($reqStr);

      if (!is_null($this->httpHandler)) {
        $response = call_user_func($this->httpHandler, $request, new \GuzzleHttp\Psr7\Response(200, [], null));
      } else {
        $response = new \GuzzleHttp\Psr7\Response(200, [], null);
      }

      fwrite($sock, \GuzzleHttp\Psr7\Message::toString($response));
    }

    fclose($sock);
  }

}