package main

import (
	"log"
)

type timeServer struct {
	listener udpListener
}

//creates a udp listener struct
func createTimeServer(clientPort int) timeServer {
	t := timeServer{}
	t.listener = createUDPListener()
	log.Println("time server created")
	return t
}
