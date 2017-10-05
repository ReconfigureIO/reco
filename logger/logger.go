package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

const timeFormat = "2006-01-02 15:04:05"

var (
	// Std is standard logger. It writes outputs to stdout.
	Std Logger = &logger{writer: os.Stdout, spinner: newSpinner(os.Stdout)}
	// Info is info logger. It writes output to stderr with prefix.
	Info Logger = &logger{writer: os.Stderr, prefix: prefix, spinner: newSpinner(os.Stderr)}
	// Error is error logger. It writes output to stderr with error prefix.
	Error Logger = &logger{writer: os.Stderr, prefix: errPrefix, spinner: newSpinner(os.Stderr)}
)

func prefix() string {
	return fmt.Sprintf("%s| ", time.Now().Format(timeFormat))
}

func errPrefix() string {
	return fmt.Sprint(prefix(), "Error: ")
}

// Logger is reco logger.
type Logger interface {
	//	Println calls Output to print to the standard logger. Arguments are handled
	//  in the manner of fmt.Println.
	Println(...interface{})
	// Printf calls Output to print to the standard logger. Arguments are handled
	// in the manner of fmt.Printf.
	Printf(string, ...interface{})
	// ShowSpinner shows a spinner to indicate progress.
	ShowSpinner(bool)
}

type logger struct {
	writer  io.Writer
	prefix  func() string
	spinner *logSpinner
	sync.Mutex
}

func (l *logger) Println(a ...interface{}) {
	l.print("", a...)
}

func (l *logger) Printf(format string, a ...interface{}) {
	l.print(format, a...)
}

func (l *logger) print(format string, a ...interface{}) {
	l.Lock()
	defer l.Unlock()

	if l.spinner.running {
		fmt.Fprintln(l.writer)
		l.showSpinner(false)
		defer l.showSpinner(true)
	} else {
		defer fmt.Fprintln(l.writer)
	}

	prefix := ""
	if l.prefix != nil {
		prefix = l.prefix()
	}

	fmt.Fprint(l.writer, prefix)
	if format == "" {
		fmt.Fprint(l.writer, a...)
	} else {
		fmt.Fprintf(l.writer, format, a...)
	}
}

func (l *logger) ShowSpinner(show bool) {
	l.Lock()
	defer l.Unlock()

	l.showSpinner(show)
}

func (l *logger) showSpinner(show bool) {
	if show {
		l.spinner.Start()
		l.spinner.running = true
	} else {
		l.spinner.Stop()
		l.spinner.running = false
	}
}

// logSpinner is a cli spinner.
type logSpinner struct {
	*spinner.Spinner
	running bool
}

// newSpinner creates a new Spinner.
func newSpinner(writer io.Writer) *logSpinner {
	sp := spinner.New([]string{"", ".", "..", "..."}, 500*time.Millisecond)
	sp.Writer = writer
	sp.Prefix = " "
	return &logSpinner{
		Spinner: sp,
	}
}
