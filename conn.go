package nighthawk

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

const (
	carReturn = "\r\n"
)

var textprotoReaderPool sync.Pool

type ConnState int

const (
	StateNew ConnState = iota
	StateActive
	StateIdle
	StateHijacked
	StateClosed
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

type conn struct {
	remoteAddr string            // network address of remote side
	rwc        net.Conn          // i/o connection
	sr         liveSwitchReader  // where the LimitReader reads from; usually the rwc
	lr         *io.LimitedReader // io.LimitReader(sr)
	buf        *bufio.ReadWriter // buffered(lr,rwc), reading from bufio->limitReader->sr->rwc
}

func (c *conn) Read(p []byte) (n int, err error) {
	// count := c.buf.Reader.Buffered()
	// log.Println("tab is not better:", count)
	//return c.rwc.Read(p)
	//return c.buf.Reader.Read(p)
	return c.sr.Read(p)
}

func (c *conn) Write(p []byte) (n int, err error) {
	return c.buf.Write(p)
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

//create a new reader from the pool
func newTextprotoReader(br *bufio.Reader) *textproto.Reader {
	if v := textprotoReaderPool.Get(); v != nil {
		tr := v.(*textproto.Reader)
		tr.R = br
		return tr
	}
	return textproto.NewReader(br)
}

//put our reader in the pool
func putTextprotoReader(r *textproto.Reader) {
	r.R = nil
	textprotoReaderPool.Put(r)
}

//reads the request and breaks it up in proper chunks
func readRequest(b *bufio.Reader) (v string, r string, h map[string][]string, buf []byte, err error) {

	tp := newTextprotoReader(b)

	var s string
	if s, err = tp.ReadLine(); err != nil {
		return "", "", nil, nil, err
	}
	defer func() {
		putTextprotoReader(tp)
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()
	verb, resource, err := parseFirstLine(s)
	if err != nil {
		log.Println("unable to read RAOP request:", err)
		return "", "", nil, nil, err
	}
	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		log.Println("unable to read RAOP mimeHeaders:", err)
		return "", "", nil, nil, err
	}
	count := b.Buffered()
	buffer, _ := b.Peek(count)

	return verb, resource, headers, buffer, nil
}

//parses and returns the verb and resource of the request
func parseFirstLine(line string) (string, string, error) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return "", "", errors.New("Invalid RTSP format")
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], nil
}

//get a header value
func getHeaderValue(headers map[string][]string, key string) string {
	if headers[key] != nil {
		return headers[key][0]
	}
	return ""
}
