package Check

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

type StatusCodeSh struct {
	ExitCodeSh int
	MutexSh    sync.Mutex
}

func (s *StatusCodeSh) Sh(scriptName string, timeSleep int) {
	for {
		// Execute the script
		cmd := exec.Command("bash", scriptName)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			log.Println(err)
		}
		s.MutexSh.Lock()
		s.ExitCodeSh = cmd.ProcessState.ExitCode()
		s.MutexSh.Unlock()

		fmt.Printf("Error code: %d\n", s.ExitCodeSh)

		// Print the output and error messages
		fmt.Println("Output:")
		fmt.Println(out.String())
		fmt.Println("Error:")
		fmt.Println(stderr.String())

		// Sleep for 5 seconds
		time.Sleep(time.Duration(timeSleep) * time.Second)
	}
}
