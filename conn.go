package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type liveSwitchReader struct {
	sync.Mutex
	r io.Reader
}

func (sr *liveSwitchReader) Read(p []byte) (n int, err error) {
	sr.Lock()
	r := sr.r
	sr.Unlock()
	return r.Read(p)
}

type ConnState int

const (
	StateNew ConnState = iota
	StateActive
	StateIdle
	StateHijacked
	StateClosed
)

type conn struct {
	remoteAddr string            // network address of remote side
	rwc        net.Conn          // i/o connection
	sr         liveSwitchReader  // where the LimitReader reads from; usually the rwc
	lr         *io.LimitedReader // io.LimitReader(sr)
	buf        *bufio.ReadWriter // buffered(lr,rwc), reading from bufio->limitReader->sr->rwc
}

// noLimit is an effective infinite upper bound for io.LimitedReader
const noLimit int64 = (1 << 63) - 1

//starts the listener and sets up the processing closure
func StartServer(port int, handler func(c *conn)) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println("error starting server:", err)
		return err
	}
	defer ln.Close()
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		rw, e := ln.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Printf("Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		//setup a connection object that handles the connection
		//this handles the RTSP protocol from interaction from here.
		c := newConn(rw)
		//c.setState(c.rwc, StateNew) // before Serve can return
		go handler(c)
	}
}

//private methods

//creates a new connection struct from the accepted socket
func newConn(rwc net.Conn) (c *conn) {
	c = new(conn)
	c.remoteAddr = rwc.RemoteAddr().String()
	c.rwc = rwc
	c.resetConn()
	return c
}

func (c *conn) resetConn() {
	c.sr = liveSwitchReader{r: c.rwc}
	c.lr = io.LimitReader(&c.sr, noLimit).(*io.LimitedReader)
	br := newBufioReader(c.lr)
	bw := newBufioWriterSize(c.rwc, 4<<10)
	c.buf = bufio.NewReadWriter(br, bw)
}

var (
	bufioReaderPool   sync.Pool
	bufioWriter2kPool sync.Pool
	bufioWriter4kPool sync.Pool
)

//helper method for creating the connection struct
func bufioWriterPool(size int) *sync.Pool {
	switch size {
	case 2 << 10:
		return &bufioWriter2kPool
	case 4 << 10:
		return &bufioWriter4kPool
	}
	return nil
}

//helper method for creating the connection struct
func putBufioReader(br *bufio.Reader) {
	br.Reset(nil)
	bufioReaderPool.Put(br)
}

//helper method for creating the connection struct
func newBufioWriterSize(w io.Writer, size int) *bufio.Writer {
	pool := bufioWriterPool(size)
	if pool != nil {
		if v := pool.Get(); v != nil {
			bw := v.(*bufio.Writer)
			bw.Reset(w)
			return bw
		}
	}
	return bufio.NewWriterSize(w, size)
}

//helper method for creating the connection struct
func newBufioReader(r io.Reader) *bufio.Reader {
	if v := bufioReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReader(r)
}
