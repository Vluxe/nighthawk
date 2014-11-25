package nighthawk

import (
	"fmt"
	"log"
	"net"
)

type updHandler func(b []byte, size int, addr *net.Addr)

type udpListener struct {
	port  int
	netLn net.UDPConn
}

var portNum = 4300

//creates a udp listener struct
func createUDPListener() udpListener {
	ln := udpListener{port: portNum} //we will switch out to port recycle thing
	portNum++
	return ln
}

//starts the listener and sets up the processing closure
func (c *udpListener) start(handler updHandler) error {
	sAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", c.port))
	if err != nil {
		log.Println("error binding UDP listener:", err)
		return err
	}
	ln, err := net.ListenUDP("udp", sAddr)
	c.netLn = *ln
	if err != nil {
		log.Println("error starting UDP listener:", err)
		return err
	}
	defer c.netLn.Close()
	for {
		buf := make([]byte, 1024)
		n, addr, err := c.netLn.ReadFrom(buf)
		if err != nil {
			log.Println("error reading from UDP listener:", err)
		}
		handler(buf, n, &addr)
	}
	return nil
}
