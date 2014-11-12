package main

import (
	"fmt"
	"log"
	"net"
)

type udpConn struct {
	port int
}

//starts the listener and sets up the processing closure
func (c *udpConn) Start(handler func(b []byte)) error {
	ln, err := net.Listen("udp", fmt.Sprintf(":%d", c.port))
	if err != nil {
		log.Println("error starting UDP listener:", err)
		return err
	}
	defer ln.Close()
	return nil
}
