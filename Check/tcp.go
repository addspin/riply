package Check

import (
	"net"
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

type StatusCodeTcp struct {
	ExitCodeTcp bool
	MyState     string
	MutexTcp    sync.Mutex
}

func (s *StatusCodeTcp) TCPPortAvailable(host string, port string, statusCodeSync *StatusCodeSync) {
	ticker := time.NewTicker(3 * time.Second)
	// s.MyState = "None"
	defer ticker.Stop()
	s.MutexTcp.Lock()
	defer s.MutexTcp.Unlock()
	for range ticker.C {
		conn, err := net.Dial("tcp", host+":"+port)
		if err != nil {
			// s.MutexTcp.Lock()
			s.ExitCodeTcp = false // port is not available
			// s.MutexTcp.Unlock()
		} else {
			conn.Close()
			// s.MutexTcp.Lock()
			s.ExitCodeTcp = true // port is available
			// s.MutexTcp.Unlock()
		}
		master := 1
		none := 0
		// Если проверка удачна и удаленный client state не Master, то становимся Master
		if s.ExitCodeTcp && statusCodeSync.ExitCodeSync == master {
			s.MyState = "Slave"
		}
		if s.ExitCodeTcp && statusCodeSync.ExitCodeSync != master {
			s.MyState = "Master"
		}
		if s.ExitCodeTcp && statusCodeSync.ExitCodeSync == none {
			s.MyState = "Master"
		}
		
		// Если проверка удачна и удаленный client state Master,
		// и мой state Master то вызываем функцйию повторно, что бы не было splitbrain
		if s.ExitCodeTcp && statusCodeSync.ExitCodeSync == master && s.MyState == "Master" {
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			s.TCPPortAvailable(host, port, statusCodeSync)
		}
		if !s.ExitCodeTcp {
			s.MyState = "Slave"
		}
	}
}
