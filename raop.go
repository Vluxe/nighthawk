package main

import (
	"encoding/hex"
	"fmt"
	"github.com/andrewtj/dnssd"
	"log"
	"net"
	"time"
)

type ConnState int

const (
	StateNew ConnState = iota
	StateActive
	StateIdle
	StateHijacked
	StateClosed
)

//starts the ROAP service
func startRAOP(hardwareAddr net.HardwareAddr, hostName string) {

	port := 5000
	name := fmt.Sprintf("%s@%s", hex.EncodeToString(hardwareAddr), hostName)
	op := dnssd.NewRegisterOp(name, "_raop._tcp", port, RegisterRAOPCallbackFunc)

	op.SetTXTPair("txtvers", "1")
	op.SetTXTPair("ch", "2")
	op.SetTXTPair("cn", "0,1")
	op.SetTXTPair("et", "0,1")
	op.SetTXTPair("sv", "false")
	op.SetTXTPair("da", "true")
	op.SetTXTPair("sr", "44100")
	op.SetTXTPair("ss", "16")
	op.SetTXTPair("pw", "false")
	op.SetTXTPair("vn", "3")
	op.SetTXTPair("tp", "TCP,UDP")
	op.SetTXTPair("md", "0,1,2")
	op.SetTXTPair("vs", "130.14")
	op.SetTXTPair("sm", "false")
	op.SetTXTPair("ek", "1")
	err := op.Start()
	if err != nil {
		log.Printf("Failed to register RAOP service: %s", err)
		return
	}
	log.Println("started RAOP service")
	go startRAOPWebServer(port)
	// later...
	//op.Stop()
}

//helper method for the ROAP service
func RegisterRAOPCallbackFunc(op *dnssd.RegisterOp, err error, add bool, name, serviceType, domain string) {
	if err != nil {
		// op is now inactive
		log.Printf("RAOP Service registration failed: %s", err)
		return
	}
	if add {
		log.Printf("RAOP Service registered as “%s“ in %s", name, domain)
	} else {
		log.Printf("RAOP Service “%s” removed from %s", name, domain)
	}
}

func startRAOPWebServer(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println("error starting RAOP server:", err)
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
				log.Printf("RAOP: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		log.Println("got a connection from: ", rw.RemoteAddr())
		//need to setup a connection object that handles the connection
		//then figure out how to handle the RTSP protocol from the data returned.

		// c, err := newConn(rw)
		// if err != nil {
		// 	continue
		// }
		// c.setState(c.rwc, StateNew) // before Serve can return
		// go c.serve()
	}
}

// func newConn(rwc net.Conn) (c *conn) {
// 	c = new(conn)
// 	c.remoteAddr = rwc.RemoteAddr().String()
// 	c.server = srv
// 	c.rwc = rwc
// 	c.sr = liveSwitchReader{r: c.rwc}
// 	c.lr = io.LimitReader(&c.sr, noLimit).(*io.LimitedReader)
// 	br := newBufioReader(c.lr)
// 	bw := newBufioWriterSize(c.rwc, 4<<10)
// 	c.buf = bufio.NewReadWriter(br, bw)
// 	return c
// }
