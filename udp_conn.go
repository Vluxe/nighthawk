package main

import (
	"fmt"
	"log"
	"net"
)

type udpListener struct {
	port int
}

//creates a udp listener struct
func createUDPListener() udpListener {
	return udpListener{port: 0} //we will switch out to port recycle thing
}

//starts the listener and sets up the processing closure
func (c *udpListener) Start(handler func(b []byte, size int)) error {
	sAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%d", c.port))
	if err != nil {
		log.Println("error binding UDP listener:", err)
		return err
	}
	ln, err := net.ListenUDP("udp", sAddr)
	if err != nil {
		log.Println("error starting UDP listener:", err)
		return err
	}
	defer ln.Close()
	for {
		buf := make([]byte, 1024)
		n, err := ln.Read(buf)
		if err != nil {
			log.Println("error reading from UDP listener:", err)
		}
		handler(buf, n)
	}
	return nil
}
