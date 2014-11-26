package nighthawk

import (
	"fmt"
	"log"
)

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
				//do stuff with video stream
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
				//resHeaders := make(map[string]string)
				if resource == "/stream.xml" {
					//respond with XML of supported features
					//How do we want to get these features?
					//c.buf.Write(s.createMirrorResponse(status, resHeaders, resData))
				} else {
					//associate with client object
					//grab important stuff out of binary plist of HTTP payload
					//start UDP time server. Client is on port 7010.
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
	log.Println("response is (minus data):", s)
	body := []byte(s + carReturn)
	if data != nil {
		body = append(body, data...)
	}
	return body
}
