package main

import (
	"bufio"
	//"bytes"
	"io"
	"log"
	"net"
	"net/textproto"
	"sync"
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

//creates a new connection struct from the accepted socket
func newConn(rwc net.Conn) (c *conn) {
	c = new(conn)
	c.remoteAddr = rwc.RemoteAddr().String()
	c.rwc = rwc
	c.sr = liveSwitchReader{r: c.rwc}
	c.lr = io.LimitReader(&c.sr, noLimit).(*io.LimitedReader)
	br := newBufioReader(c.lr)
	bw := newBufioWriterSize(c.rwc, 4<<10)
	c.buf = bufio.NewReadWriter(br, bw)
	return c
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

//starts the connection processing
func (c *conn) serve() {
	log.Println("got a connection from: ", c.rwc.RemoteAddr())
	ReadRequest(c.buf.Reader)
	c.buf.Writer.Write([]byte("hi"))
	c.rwc.Close()
}

var textprotoReaderPool sync.Pool

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

//read the first line of the request
func ReadRequest(b *bufio.Reader) (err error) {

	tp := newTextprotoReader(b)

	var s string
	if s, err = tp.ReadLine(); err != nil {
		return err
	}
	defer func() {
		putTextprotoReader(tp)
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()
	log.Println("first: ", s)
	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		log.Println("unable to read mimeHeaders:", err)
	}
	log.Println("headers are: ", headers)
	count := b.Buffered()
	log.Println("buffer contains:", count)
	buffer, _ := b.Peek(count)
	log.Println("all the body len: ", len(buffer))
	//buffer := new(bytes.Buffer)
	//io.Copy(buffer, b)
	//log.Println("all the body len: ", len(buffer.Bytes()))
	return nil
}
