package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	// Std is standard logger. It writes outputs to stdout.
	Std Logger = &logger{writer: os.Stdout}
	// Info is info logger. It writes output to stderr with prefix.
	Info Logger = &logger{writer: os.Stderr, prefix: prefix}
	// Error is error logger. It writes output to stderr with error prefix.
	Error Logger = &logger{writer: os.Stderr, prefix: errPrefix}
)

func prefix() string {
	return ""
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
}

type logger struct {
	writer io.Writer
	prefix func() string
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

	defer fmt.Fprintln(l.writer)

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
