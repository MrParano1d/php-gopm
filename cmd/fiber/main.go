package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mrparano1d/php-gopm/pkg/config"
	"github.com/mrparano1d/php-gopm/pkg/process"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
)

var profile = flag.Bool("profile", false, "write profile to `file`")
var scriptPath = flag.String("script", "./scripts/app.php", "the php script that should be run")

func main() {

	flag.Parse()

	if *profile {
		cf, err := os.Create("cpu.pprof")
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(cf); err != nil {
			panic(fmt.Errorf("failed to start cpu profile: %v", err))
		}

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			mf, err := os.Create("mem.pprof")
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}

			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(mf); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
			pprof.StopCPUProfile()

			if err := mf.Close(); err != nil {
				log.Printf("failed to close memory pprof file: %v\n", err)
			}
			if err := cf.Close(); err != nil {
				log.Printf("failed to close cpu pprof file: %v\n", err)
			}
			os.Exit(1)
		}()
	}

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
