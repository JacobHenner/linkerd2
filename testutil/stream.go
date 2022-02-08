package testutil

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Stream provides the ability of read the output of an executing process while
// it is still running
type Stream struct {
	cmd *exec.Cmd
	out io.ReadCloser
}

// Stop closes the stream and kills the process
func (s *Stream) Stop() {
	if err := s.out.Close(); err != nil {
		logrus.WithError(err).Error("Failed to close stream")
	}
	if err := s.cmd.Process.Kill(); err != nil {
		logrus.WithError(err).Error("Failed to kill process")
	}
}

// ReadUntil reads from the process output until specified number of lines has
// been reached, or until a timeout
func (s *Stream) ReadUntil(lineCount int, timeout time.Duration) ([]string, error) {
	output := make([]string, 0)
	lines := make(chan string)
	timeoutAfter := time.After(timeout)
	scanner := bufio.NewScanner(s.out)
	stopSignal := false

	go func() {
		for scanner.Scan() {
			lines <- scanner.Text()

			if stopSignal {
				close(lines)
				return
			}
		}
	}()

	for {
		select {
		case <-timeoutAfter:
			stopSignal = true
			return output, fmt.Errorf("cmd [%s] Timed out trying to read %d lines", strings.Join(s.cmd.Args, " "), lineCount)
		case line := <-lines:
			output = append(output, line)
			if len(output) >= lineCount {
				stopSignal = true
				return output, nil
			}
		}
	}
}
