package proxy

import (
	"io"
	"log"
	"net"
	"sync"
)

type Info struct {
	// Listen is the address proxy server is listening on.
	// This is nil unless the proxy server is running.
	Listen *net.TCPAddr
	// Remote is the address proxy server is forwarding to.
	Remote *net.TCPAddr

	// Error is the error that causes the proxy server to terminate abruptly.
	Error error
}

// Proxy is a TCP proxy.
type Proxy interface {
	// Start starts the proxy server.
	Start() error

	// Stop stops the proxy server.
	Stop() error

	// Wait waits for the proxy server to terminate.
	Wait()

	// Info gives information about the proxy connection.
	Info() Info
}

var _ Proxy = &proxyImpl{}

type proxyImpl struct {
	info     *Info
	listener net.Listener
	wait     chan struct{}
}

// New creates a new proxy server that proxies connection to addr.
func New(addr string) (Proxy, error) {
	remote, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &proxyImpl{
		info: &Info{
			Remote: remote,
		},
	}, nil
}

func (p *proxyImpl) Start() error {
	if p.info.Listen == nil {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			return err
		}
		p.listener = l
		p.info.Listen = l.Addr().(*net.TCPAddr)
		p.wait = make(chan struct{})
	}

	go p.listen()
	return nil
}

func (p *proxyImpl) listen() {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			log.Println(err)
			if _, ok := err.(*net.OpError); ok {
				p.info.Error = err
				break
			}
		}

		go p.handle(conn)
	}

	close(p.wait)
}

func (p *proxyImpl) handle(conn net.Conn) {
	client, err := net.DialTCP("tcp", nil, p.info.Remote)
	if err != nil {
		log.Println(err)
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
		return
	}
	var w sync.WaitGroup
	w.Add(2)
	go func() {
		defer conn.Close()
		defer client.Close()
		io.Copy(client, conn)
		w.Done()
	}()
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(conn, client)
		w.Done()
	}()
	w.Wait()
}

func (p *proxyImpl) Stop() error {
	if p.info.Listen == nil {
		return nil
	}
	p.info.Listen = nil
	return p.listener.Close()
}

func (p *proxyImpl) Info() Info {
	return *p.info
}

func (p *proxyImpl) Wait() {
	<-p.wait
}
