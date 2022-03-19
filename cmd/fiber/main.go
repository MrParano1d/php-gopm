package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mrparano1d/php-gopm/pkg/config"
	"github.com/mrparano1d/php-gopm/pkg/process"
)

var scriptPath = flag.String("script", "./scripts/app.php", "the php script that should be run")

func main() {

	flag.Parse()

	manager := process.NewManager(&config.Config{
		ScriptPath: *scriptPath,
	})

	l, err := net.Listen("tcp", "localhost:13337")
	if err != nil {
		panic(fmt.Errorf("failed to listen to :13337: %v", err))
	}

	go func() {
		if err := manager.Start(l); err != nil {
			panic(err)
		}
	}()

	app := fiber.New()

	app.Get("*", func(c *fiber.Ctx) error {

		res, err := manager.Request(c.Request().String())
		if err != nil {
			return err
		}

		strReader := strings.NewReader(res)
		reader := bufio.NewReader(strReader)

		resp, err := http.ReadResponse(reader, nil)
		if err != nil {
			return err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("failed to close response body: %v\n", err)
			}
		}()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		for k, vls := range resp.Header {
			c.Set(k, strings.Join(vls, ","))
		}

		c.Status(resp.StatusCode)

		return c.Send(bodyBytes)
	})

	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("failed to start app: %v", err)
	}
}
