package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type workerConfig struct {
	httpAddr string
	tcpAddr  string
}

type workerInfo struct {
	conn     net.Conn
	httpAddr string
	tcpAddr  string
	alive    bool
}

type Master struct {
	workers map[string]*workerInfo
	volumes map[string]string
	mu      sync.Mutex
	cfgPath string
}

func newMaster(cfgPath string) (*Master, error) {
	m := &Master{}
	m.cfgPath = cfgPath

	m.workers = make(map[string]*workerInfo)
	m.volumes = make(map[string]string)

	w, err := os.ReadFile("./workers.json")
	if err != nil {
		log.Fatal("File not found")
	}
	payload := &workerConfig{}
	err = json.Unmarshal(w, payload)
	if err != nil {
		log.Printf("Worker Config Unmarshalling error: %v", err)
	}

	return m, nil
}

func startMaster(cfgPath string) error {
	m, err := newMaster(cfgPath)
	if err != nil {
		log.Fatalf("Failed to start the Master server: %v", err)
	}

}

func (m *Master) selectWorker() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for addr, w := range m.workers {
		if w.alive {
			return addr, nil
		}
	}
	return "", errors.New("no alive workers")
}

func (m *Master) registerWorker(httpAddr, tcpAddr string) error {
	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		return errors.New("dial error when attempting register worker")
	}
	r := encodeRegister(httpAddr)
	if _, err = conn.Write(r); err != nil {
		conn.Close()
		log.Printf("Failed to write register request: %v", err)
	}
	msgType, _, err := decodeMessage(conn)
	if err != nil {
		conn.Close()
		log.Printf("Failed to decode register response: %v", err)
	}
	if msgType != successMsg {
		conn.Close()
		log.Printf("Invalid decode: %v", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.workers[httpAddr] = &workerInfo{
		conn:     conn,
		httpAddr: httpAddr,
		tcpAddr:  tcpAddr,
		alive:    true,
	}

	go m.healthCheck(httpAddr, tcpAddr)

	return nil
}

func (m *Master) healthCheck(httpAddr, tcpAddr string) {
	for {
		time.Sleep(time.Second * 10)
		m.mu.Lock()
		c := m.workers[httpAddr].conn
		p := encodePing()
		if _, err := c.Write(p); err != nil {
			m.workers[httpAddr].alive = false
			err := m.registerWorker(httpAddr, tcpAddr)
			if err != nil {
				log.Printf("Failed to re-register worker after PING failure: %v", err)
			}
		}
		msgType, payload, err := decodeMessage(c)
		if msgType == successMsg && string(payload) == "OK" {
			m.mu.Lock()
			m.workers[httpAddr].alive = true
		}
		if err != nil {
			m.mu.Lock()
			m.workers[httpAddr].alive = true
			err := m.registerWorker(httpAddr, tcpAddr)
			if err != nil {
				log.Printf("Failed to re-register worker after PONG failure: %v", err)
			}
		}
		log.Printf("Worker current status: %v", m.workers[httpAddr].alive)
	}
}
