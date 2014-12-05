package nighthawk

import (
	//"fmt"
	"log"
	"net"
)

//The client struct is used to encapsulate a device (OS X or iOS).
type Client struct {
	Name          string      //the name that the client is reporting
	RTSPUrl       string      //the RTSP stream URL
	deviceIP      string      //The IP of the client
	rasAesKey     []byte      //crypto keys
	aesivKey      []byte      //crypto keys
	deviceID      string      //the Apple device ID a (hex value of the client)
	serverLn      udpListener //the server udp listener
	controlLn     udpListener //the control udp listener
	timeSvr       timeServer  //the udp time server
	mirrorTimeSvr timeServer  //the udp time server of the screen mirroring
}

//setup the UDP ports. port is the timing port for the timeServer
//returns the 3 udp ports created. server port, then control port, then the time server port
func (c *Client) setup(timePort int, serverHandler func(b []byte, size int)) (int, int, int) {
	c.serverLn = createUDPListener()
	go c.serverLn.start(func(b []byte, size int, addr *net.Addr) {
		//decrypt audio packets
		serverHandler(b, size)
	})
	c.controlLn = createUDPListener()
	go c.controlLn.start(func(b []byte, size int, addr *net.Addr) {
		//process Sync packets
		log.Println("Sync packet!")
	})
	c.timeSvr = createTimeServer(timePort, c.deviceIP)
	return c.serverLn.port, c.controlLn.port, c.timeSvr.listener.port
}

//start the audio stream by starting client's time server
func (c *Client) start() {
	go c.timeSvr.start()
}

//start the audio stream
func (c *Client) stop() {
	c.timeSvr.stop()
}

//stop the udp listeners and cleanup things
func (c *Client) teardown() {
	c.mirrorTimeSvr.stop() //this client is done!
}

//start the mirroring stream by starting client's time server
func (c *Client) startMirror() {
	c.mirrorTimeSvr = createMirrorTimeServer(c.deviceIP)
	go c.mirrorTimeSvr.start()
}
