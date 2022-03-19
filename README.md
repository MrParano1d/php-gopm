# PHP-GOPM

!WIP!

**Warning** 10% of the traffic is currently swallowed by the http server. :(

PHP-GOPM is a php process manager written it go.

Currently, it's based on TCP client sockets in the php code and a TCP server in go.

It starts a http server which sends the http request to a php process and returns the php response as a http response.

In `./scripts` is the app.php with a small server library that handles the go/php communication.

The PHP request/response are based on the psr7 standard.

The PHP script is not interpreted, but started once with the `php` command from the host system. Changes to the script are not currently recognized, but the server must be restarted.

## Why?

Our product at work can be addressed by any programming language. However, since PHP is an interpreted language, we needed a way that we could pass our asynchronous lifecycles from Go to PHP.

So we continue to run the PHP "SDK" for our product via Go, but developers can implement the business logic in PHP code.

## Performance

At first glance, PHP-GOPM makes a good impression. In our tests PHP-GOPM was even faster than PHP-FPM.

However, if you don't need Go code at all on the server side, then you should rather read about projects like "swoole" and use them.

## TODOs

- [ ] Map PHP Process to TCP connection
- [ ] Process Manager Configuration
- [ ] Reload PHP workers when script file changes
- [ ] Docs
- [ ] Unix Sockets
- [x] Http Calls - see `./scripts`
- [ ] CLI Calls
