package Check

import (
	"bytes"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3/client"
)

type StatusCodeSync struct {
	ExitCodeSync int
	BodyData     string
	SyncState    string
	MutexSync    sync.Mutex
}

func (s *StatusCodeSync) Sync(host, port string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	cc := client.New()

	for range ticker.C {

		resp, err := cc.Get("http://" + host + ":" + port + "/state")
		if err != nil {
			log.Println(err)
			s.ExitCodeSync = 0
			continue
		}

		// Read the response body
		s.MutexSync.Lock()
		// defer s.MutexSync.Unlock()
		bodyBytes := resp.Body()

		// log.Println("BODY:", bodyBytes)
		bodyReader := bytes.NewReader(bodyBytes)
		body, err := io.ReadAll(bodyReader)

		if err != nil {
			log.Println(err)

		}
		// log.Println("SyncState: ", s.SyncState)
		s.BodyData = string(body)
		if s.BodyData != "Master" && s.BodyData != "Slave" {
			// s.ExitCodeSync = false // port is not available
			s.ExitCodeSync = 0
			// s.SyncState = "Slave"
			s.SyncState = "None"
		}
		if s.BodyData == "Slave" {
			// s.ExitCodeSync = true
			s.ExitCodeSync = 2 //
			s.SyncState = "Slave"
		}
		if s.BodyData == "Master" {
			// s.ExitCodeSync = true //
			s.ExitCodeSync = 1
			s.SyncState = "Master"
		}
		resp.Close()
		s.MutexSync.Unlock()
	}
}
