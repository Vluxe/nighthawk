package nighthawk

import (
	"fmt"
	"log"
)

type MirrorFeatures struct {
	Height      int     //height of your screen
	Width       int     //width of your screen
	Overscanned bool    //is the display overscanned?
	RefreshRate float32 //refresh rate 60 Hz (1/60) 0.016666666666666666
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
				//log.Println("got mirror video packet!")
				//do stuff with and parse video stream
				//add interface stuff here too
			} else {
				verb, resource, headers, data, err := readRequest(c.buf.Reader)
				log.Println("verb:", verb)
				log.Println("headers:", headers)
				log.Println("data:", data)
				if err != nil {
					log.Println("error process mirror server request:", err)
					return
				}
				resHeaders := make(map[string]string)
				if resource == "/stream.xml" {
					f := s.delegate.SupportedMirrorFeatures()
					d := s.createFeaturesResponse(f)
					c.buf.Write(s.createMirrorResponse(true, resHeaders, d))
				} else {
					log.Println("Got the second mirror stream!")
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
func (server *airServer) createMirrorResponse(success bool, headers map[string]string, data []byte) []byte {
	s := HTTPProtocolType
	if success {
		s += " 200 OK" + carReturn
		if data != nil {
			s += fmt.Sprintf("Content-Type: text/x-apple-plist+xml%s", carReturn)
			s += fmt.Sprintf("Content-Length: %d%s", len(data), carReturn)
		}
		for key, val := range headers {
			s += fmt.Sprintf("%s: %s%s", key, val, carReturn)
		}
	} else {
		s += " 400 Bad Request" + carReturn
	}
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
