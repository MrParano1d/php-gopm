package process

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

type PHPProcess struct {
	scriptPath string
	idle       bool

	stopRetry bool

	conn net.Conn
}

func NewPHPProcess(scriptPath string) *PHPProcess {
	return &PHPProcess{
		scriptPath: scriptPath,
		idle:       true,
		stopRetry:  false,
	}
}

func (p *PHPProcess) KeepAlive(c net.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if c != nil {
				_, err := c.Write([]byte("ping"))
				if err != nil {
					log.Printf("failed to send ping message to php process: %v\n", err)
				}
			}
		}
	}
}

func (p *PHPProcess) IsIdle() bool {
	return p.idle
}

func (p *PHPProcess) Connect(c net.Conn) {

	go p.KeepAlive(c)

	p.conn = c

}

func (p *PHPProcess) Disconnect() {
	p.conn = nil
}

func (p *PHPProcess) Handle(request string) (string, error) {
	p.idle = false
	defer func() {
		p.idle = true
	}()

	if p.conn == nil {
		return "", errors.New("process not connected")
	}

	if _, err := p.conn.Write([]byte(request)); err != nil {
		return "", fmt.Errorf("failed to write to php process: %v", err)
	}

	buffer := make([]byte, 1024)
	_, err := p.conn.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read from worker stream: %v", err)
	}

	return string(bytes.Trim(buffer, "\x00")), nil
}

func (p *PHPProcess) startProcess() error {
	p.idle = true
	if p.stopRetry == false {
		defer func() {
			if r := recover(); r != nil {
				p.Disconnect()
				log.Println("php process died: restarting")
				if err := p.startProcess(); err != nil {
					p.stopRetry = true
				}
			}
		}()
	}
	cmd := exec.Command("php", p.scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(fmt.Errorf("failed to run php script %s: %v", p.scriptPath, err))
	}
	return nil
}

func (p *PHPProcess) Run() error {

	if err := p.startProcess(); err != nil {
		return err
	}

	return nil
}
