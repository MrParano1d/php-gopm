package process

import (
	"errors"
	"fmt"
	"log"
	"net"
	"runtime"
	"time"

	"github.com/mrparano1d/php-gopm/pkg/config"
)

type Manager struct {
	numWorkers      int
	config          *config.Config
	workers         []*PHPProcess
	connectionQueue chan net.Conn
	idleWorker      chan *PHPProcess
}

func NewManager(config *config.Config) *Manager {
	var numWorkers int
	if config.NumWorkers == 0 {
		numWorkers = runtime.NumCPU()
	} else {
		numWorkers = config.NumWorkers
	}

	if config.RequestTimeout == 0 {
		config.RequestTimeout = 10000
	}

	return &Manager{
		numWorkers:      numWorkers,
		config:          config,
		workers:         make([]*PHPProcess, numWorkers),
		connectionQueue: make(chan net.Conn, numWorkers),
		idleWorker:      make(chan *PHPProcess, 1),
	}
}

func (m *Manager) SpawnWorkers() error {

	log.Printf("spawing %d workers\n", m.numWorkers)

	for i := 0; i < m.numWorkers; i++ {
		m.workers[i] = NewPHPProcess(m.config.ScriptPath)
		go func(w *PHPProcess) {
			if err := w.Run(); err != nil {
				log.Printf("failed to run worker: %v\n", err)
			}
		}(m.workers[i])
	}

	return nil
}

func (m *Manager) Listen(l net.Listener) {
	for {
		fd, err := l.Accept()
		if err != nil {
			log.Printf("failed to accept tcp request: %v\n", err)
			continue
		}

		m.connectionQueue <- fd
	}
}

func (m *Manager) handleConnectionQueue() {
	for {
		select {
		case c := <-m.connectionQueue:
			worker := <-m.idleWorker
			worker.Connect(c)
		}
	}
}

func (m *Manager) Request(req string) (string, error) {
	h := NewPHPHandler(req)

	ticker := time.NewTicker(time.Duration(m.config.RequestTimeout) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			return "", errors.New("request timeout")
		case worker := <-m.idleWorker:
			if res, err := worker.Handle(h.Request); err != nil {
				return "", err
			} else {
				return res, nil
			}
		default:
			fmt.Println(time.Now().String(), "waiting for idle worker")
		}
	}
}

func (m *Manager) getIdleWorker() {
	for {
		for _, w := range m.workers {
			if w.IsIdle() {
				m.idleWorker <- w
			}
		}
	}
}

func (m *Manager) Start(l net.Listener) error {
	errC := make(chan error)

	go m.Listen(l)

	if err := m.SpawnWorkers(); err != nil {
		return fmt.Errorf("failed to spawn workers: %v", err)
	}

	go m.getIdleWorker()

	go m.handleConnectionQueue()

	return <-errC
}
