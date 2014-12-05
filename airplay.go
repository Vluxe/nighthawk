package nighthawk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	//"io"
	"log"
)

type MirrorFeatures struct {
	Height      int     //height of your screen
	Width       int     //width of your screen
	Overscanned bool    //is the display overscanned?
	RefreshRate float32 //refresh rate 60 Hz (1/60) 0.016666666666666666
}

type streamPacketHeader struct {
	PayloadSize  int32
	PayloadType  int16
	PayloadG     int16 // not sure what this does.
	NTPTimestamp int64
	Ignore       [112]byte //ignore the rest of header.
}

const (
	airplayPort      = 7000
	mirroringPort    = 7100
	HTTPProtocolType = "HTTP/1.1"
	//rstAvoidanceDelay = 500 * time.Millisecond
)

// startAirplay starts the airplay server.
func (s *airServer) startAirplay() {
	log.Println("started Airplay service")
	// I believe there is some sort of generic http server that handles video streaming on port 7000.
	s.startMirroringWebServer(mirroringPort) //for screen mirroring
}

func (s *airServer) startMirroringWebServer(port int) {
	StartServer(port, func(c *conn) {
		log.Println("got a Mirror connection from: ", c.rwc.RemoteAddr())
		isStream := false
		for {
			if isStream {
				s.handleVideoStream(c)
				//add interface stuff here too
			} else {
				verb, resource, headers, data, err := readRequest(c.buf.Reader)
				log.Println("resource:", resource)
				log.Println("verb:", verb)
				log.Println("headers:", headers)
				if err != nil {
					log.Println("error process mirror server request:", err)
					return
				}
				resHeaders := make(map[string]string)
				resHeaders["User-Agent"] = "AirPlay/215.10"
				if resource == "/stream.xml" {
					f := s.delegate.SupportedMirrorFeatures()
					d := s.createFeaturesResponse(f)
					c.buf.Write(s.createMirrorResponse(true, true, resHeaders, d))
				} else if resource == "/fp-setup" {
					resData := s.handleFairPlay(resHeaders, data)
					c.buf.Write(s.createMirrorResponse(true, false, resHeaders, resData))
				} else {
					//log.Println("Got the second mirror stream!")
					host := s.getClientIP(c)
					client := s.clients[host]
					if client != nil {
						//grab important stuff out of binary plist of HTTP payload
						client.startMirror()
					}
					isStream = true //set if the stream has change from HTTP to video
				}
			}
			// if !status {
			// 	c.rwc.Close()
			// }
			c.buf.Flush()
			c.resetConn()
		}
	})
}

//creates a response to send back to the client
func (server *airServer) createMirrorResponse(success bool, isPlist bool, headers map[string]string, data []byte) []byte {
	s := HTTPProtocolType
	if success {
		s += " 200 OK" + carReturn
		if data != nil {
			if isPlist {
				s += fmt.Sprintf("Content-Type: text/x-apple-plist+xml%s", carReturn)
			} else {
				s += fmt.Sprintf("Content-Type: application/octet-stream%s", carReturn)
			}
			s += fmt.Sprintf("Content-Length: %d%s", len(data), carReturn)
		}
		for key, val := range headers {
			s += fmt.Sprintf("%s: %s%s", key, val, carReturn)
		}
	} else {
		s += " 400 Bad Request" + carReturn
	}
	log.Println("response is (minus data):", s)
	body := []byte(s + carReturn)
	if data != nil {
		body = append(body, data...)
	}
	return body
}

//create a XML response of the MirrorFeatures
func (server *airServer) createFeaturesResponse(f MirrorFeatures) []byte {
	scanned := "<true/>"
	if !f.Overscanned {
		scanned = "<false/>"
	}
	str := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
 "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
 <dict>
  <key>height</key>
  <integer>%d</integer>
  <key>overscanned</key>
  %s
  <key>refreshRate</key>
  <real>%f</real>
  <key>version</key>
  <string>150.33</string>
  <key>width</key>
  <integer>%d</integer>
 </dict>
</plist>`, f.Height, scanned, f.RefreshRate, f.Width)
	return []byte(str)
}

func (server *airServer) handleVideoStream(c *conn) {
	buffer := make([]byte, 128)
	for {
		n, err := c.Read(buffer)
		if err != nil { //&& err != io.EOF
			log.Println(err)
		}
		if n <= 0 {
			break
		}
		var header streamPacketHeader
		buf := bytes.NewReader(buffer)
		err = binary.Read(buf, binary.LittleEndian, &header)
		if err != nil {
			fmt.Println("binary.Read failed:", err)
		}
		log.Println("Stream Packet Header.")
		log.Println(header.PayloadType)
		log.Println(header.PayloadSize)
	}
}
