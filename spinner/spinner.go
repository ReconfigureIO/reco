package spinner

import (
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
)

// Spinner struct to hold the provided options
type Spinner struct {
	Delay      time.Duration // Delay is the speed of the indicator
	chars      []string      // chars holds the chosen character set
	Prefix     string        // Prefix is the text preppended to the indicator
	lastOutput string        // last character(set) written
	lock       *sync.RWMutex //
	Writer     io.Writer     // to make testing better, exported so users have access
	active     bool          // active holds the state of the spinner
	stopChan   chan struct{} // stopChan is a channel used to stop the indicator
}

// New provides a pointer to an instance of Spinner with the supplied options
func New(cs []string, d time.Duration) *Spinner {
	return &Spinner{
		Delay:    d,
		chars:    cs,
		color:    color.New(color.FgWhite).SprintFunc(),
		lock:     &sync.RWMutex{},
		Writer:   color.Output,
		active:   false,
		stopChan: make(chan struct{}, 1),
	}
}

// Start will start the indicator
func (s *Spinner) Start() {
	if s.active {
		return
	}
	s.active = true

	go func() {
		for {
			for i := 0; i < len(s.chars); i++ {
				select {
				case <-s.stopChan:
					return
				default:
					s.lock.Lock()
					outColor := fmt.Sprintf("%s%s%s ", s.Prefix, s.color(s.chars[i]), s.Suffix)
					outPlain := fmt.Sprintf("%s%s%s ", s.Prefix, s.chars[i], s.Suffix)
					fmt.Fprint(s.Writer, outColor)
					delay := s.Delay
					s.lock.Unlock()

					time.Sleep(delay)
				}
			}
		}
	}()
}

// Stop stops the indicator
func (s *Spinner) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.active {
		s.active = false
		s.stopChan <- struct{}{}
	}
}

// Restart will stop and start the indicator
func (s *Spinner) Restart() {
	s.Stop()
	s.Start()
}
