package main

import (
//"fmt"
//"log"
)

//The client struct is used to encapsulate a device (OS X or iOS).
type Client struct {
	Name      string      //the name that the client is reporting
	RTSPUrl   string      //the RTSP stream URL
	rasAesKey []byte      //crypto keys
	aesivKey  []byte      //crypto keys
	deviceID  string      //the Apple device ID a (hex value of the client)
	serverLn  udpListener //the server udp listener
	controlLn udpListener //the control udp listener
	timeSvr   timeServer  //the udp time server
}

//setup the UDP ports. port is the timing port for the timeServer
//returns the 3 udp ports created. server port, then control port, then the time server port
func (c *Client) setup(timePort int) (int, int, int) {
	c.serverLn = createUDPListener()
	c.controlLn = createUDPListener()
	c.timeSvr = createTimeServer(timePort)
	return c.serverLn.port, c.controlLn.port, c.timeSvr.listener.port
}

//start the audio stream
func (c *Client) record() {

}
