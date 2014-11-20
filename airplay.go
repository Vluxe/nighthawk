package nighthawk

import (
	"log"
)

const (
	airplayPort   = 7000
	mirroringPort = 7100
)

// startAirplay starts the airplay server.
func (s *airServer) startAirplay() {
	log.Println("started Airplay service")
	// I believe there is some sort of generic http server that handles video streaming on port 7000.
	startMirroringWebServer(mirroringPort) //for screen mirroring
}

func startMirroringWebServer(port int) {
	StartServer(port, func(c *conn) {
		log.Println("got a Mirror connection from: ", c.rwc.RemoteAddr())
		//s.readRequest(c.buf.Reader)
		c.buf.Write([]byte("OK"))
	})
}
