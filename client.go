package main

import (
//"fmt"
//"log"
)

//The client struct is used to encapsulate a device (OS X or iOS).
type Client struct {
	Name     string //the name that the client is reporting
	RTSPUrl  string //the RTSP stream URL
	fpaeskey string //crypto keys
	aesiv    string //crypto keys
	deviceID string //the Apple device ID a (hex value of the client)
}

//setup the UDP ports
func (c *Client) setup() {

}

//start the audio stream
func (c *Client) record() {

}
