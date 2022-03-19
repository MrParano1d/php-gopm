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

	conn net.Conn
}

func NewPHPProcess(scriptPath string) *PHPProcess {
	return &PHPProcess{
		scriptPath: scriptPath,
		idle:       true,
	}
}

func (p *PHPProcess) KeepAlive(c net.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			_, err := c.Write([]byte("ping"))
			if err != nil {
				log.Printf("failed to send ping message to php process: %v\n", err)
			}
		}
	}
}

func (p *PHPProcess) IsIdle() bool {
	return p.idle
}

func (p *PHPProcess) Connect(c net.Conn) {

	go p.KeepAlive(c)

	log.Println("connecting")
	p.conn = c

}

func (p *PHPProcess) Handle(request string) (string, error) {
	p.idle = false

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
	//response, err := io.ReadAll(bufio.NewReader(p.conn))

	p.idle = true
	return string(bytes.Trim(buffer, "\x00")), nil
}

func (p *PHPProcess) Run() error {
	cmd := exec.Command("php", p.scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run php script %s: %v", p.scriptPath, err)
	}

	return nil
}
