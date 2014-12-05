package nighthawk

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	RTSPProtocolType = "RTSP/1.0"
	raopPort         = 5000
	//rstAvoidanceDelay = 500 * time.Millisecond
)

// startRAOPServer starts the RTSP/RAOP server.
func (s *airServer) startRAOPServer() {
	StartServer(raopPort, func(c *conn) {
		log.Println("got a RAOP connection from: ", c.rwc.RemoteAddr())
		for {
			verb, resource, headers, data, err := readRequest(c.buf.Reader)
			if err != nil {
				return
			}
			resHeaders := make(map[string]string)
			resHeaders["Server"] = "AirTunes/150.33"
			key := "Cseq"
			if headers[key] != nil {
				resHeaders[key] = headers[key][0]
			}
			resData, status := s.processRTSPRequest(c, verb, resource, headers, resHeaders, data)
			c.buf.Write(s.createRTSPResponse(status, resHeaders, resData))
			c.buf.Flush()
			c.resetConn()
			if !status || verb == "TEARDOWN" {
				c.rwc.Close()
			}
		}
	})
	log.Println("RAOP server finished...?")
}

//creates a response to send back to the client
func (server *airServer) createRTSPResponse(success bool, headers map[string]string, data []byte) []byte {
	s := RTSPProtocolType
	if success {
		s += " 200 OK" + carReturn
		if data != nil {
			s += fmt.Sprintf("Content-Type: application/octet-stream%s", carReturn)
			s += fmt.Sprintf("Content-Length: %d%s", len(data), carReturn)
		}
		for key, val := range headers {
			s += fmt.Sprintf("%s: %s%s", key, val, carReturn)
		}
	} else {
		s += " 400 Bad Request" + carReturn
	}
	//log.Println("response is (minus data):", s)
	body := []byte(s + carReturn)
	if data != nil {
		body = append(body, data...)
	}
	return body
}

//processes the request by dispatching to the proper method for each response
func (s *airServer) processRTSPRequest(c *conn, verb, resource string, headers map[string][]string, resHeaders map[string]string, data []byte) ([]byte, bool) {
	log.Println("resource is:", resource)
	log.Println("verb is:", verb)
	if verb == "POST" && resource == "/fp-setup" {
		return s.handleFairPlay(resHeaders, data), true
	} else if verb == "POST" && resource == "/auth-setup" {
		return nil, true
	} else if verb == "OPTIONS" {
		return s.handleOptions(resource, headers, resHeaders)
	} else if verb == "ANNOUNCE" {
		status := s.handleAnnounce(c, resource, headers, data)
		return nil, status
	} else if verb == "SETUP" {
		return s.handleSetup(c, resource, headers, resHeaders)
	} else if verb == "RECORD" {
		return s.handleRecord(c, resource, headers, resHeaders)
	} else if verb == "SET_PARAMETER" {
		return s.handleSetParameters(c, resource, headers, resHeaders)
	} else if verb == "FLUSH" {
		return s.handleFlush(c, resource, headers, resHeaders)
	} else if verb == "TEARDOWN" {
		return s.handleTeardown(c, resource, headers, resHeaders)
	} else if verb == "GET_PARAMETER" {
		return nil, true //this verb doesn't do anything or even get called, but just to be safe
	}
	log.Println("RTSP: not sure how to handle this...")
	return nil, false
}

func (s *airServer) getClientIP(con *conn) string {
	host, _, _ := net.SplitHostPort(con.rwc.RemoteAddr().String())
	return host
}

//process fair play requests
func (s *airServer) handleFairPlay(headers map[string]string, data []byte) []byte {
	if data[6] == 1 {
		log.Println("fairplay type 1")
		return []byte{0x46, 0x50, 0x4c, 0x59, 0x03, 0x01, 0x02, 0x00, 0x00, 0x00, 0x00, 0x82, 0x02, 0x00,
			0x65, 0x2b, 0x94, 0x81, 0x14, 0xd4, 0xc3, 0x86, 0x6f, 0x20, 0xd5, 0x9b, 0x1c, 0x48, 0xc5, 0x7f,
			0x51, 0x1d, 0x32, 0x31, 0x0f, 0x5f, 0xf3, 0x92, 0x1d, 0x76, 0x37, 0xd5, 0xaa, 0xb5, 0x70, 0x40,
			0xa3, 0xac, 0x0f, 0x54, 0x87, 0x78, 0xb8, 0xd3, 0xc3, 0xa9, 0xe5, 0x0e, 0x1e, 0x6e, 0x24, 0xe8,
			0xa1, 0x43, 0x75, 0x3d, 0xd5, 0x32, 0xad, 0xf3, 0xb8, 0xc8, 0x35, 0x71, 0xeb, 0x7c, 0x81, 0x3c,
			0x67, 0x0d, 0xdd, 0x4d, 0xcd, 0x20, 0x76, 0xc1, 0x70, 0xe3, 0x9a, 0xc6, 0x1e, 0xf7, 0x1d, 0x5a,
			0x58, 0x09, 0x62, 0x4f, 0x14, 0x69, 0x9c, 0x07, 0xd8, 0x12, 0x5f, 0x22, 0x28, 0x29, 0x72, 0xfc,
			0x66, 0xf8, 0x46, 0x46, 0xda, 0x3d, 0x10, 0xcf, 0x05, 0xc8, 0x99, 0x1c, 0x3a, 0x57, 0xa9, 0xc2,
			0x0e, 0xe9, 0x12, 0x91, 0x8c, 0x96, 0xc3, 0x43, 0x2a, 0x0b, 0xf8, 0x71, 0xde, 0x97, 0xd4, 0xbf}
		// return []byte{0x46, 0x50, 0x4c, 0x59, 0x02, 0x01, 0x02, 0x00, 0x00, 0x00, 0x00, 0x82,
		// 	0x02, 0x02, 0x2f, 0x7b, 0x69, 0xe6, 0xb2, 0x7e, 0xbb, 0xf0, 0x68, 0x5f, 0x98, 0x54, 0x7f, 0x37,
		// 	0xce, 0xcf, 0x87, 0x06, 0x99, 0x6e, 0x7e, 0x6b, 0x0f, 0xb2, 0xfa, 0x71, 0x20, 0x53, 0xe3, 0x94,
		// 	0x83, 0xda, 0x22, 0xc7, 0x83, 0xa0, 0x72, 0x40, 0x4d, 0xdd, 0x41, 0xaa, 0x3d, 0x4c, 0x6e, 0x30,
		// 	0x22, 0x55, 0xaa, 0xa2, 0xda, 0x1e, 0xb4, 0x77, 0x83, 0x8c, 0x79, 0xd5, 0x65, 0x17, 0xc3, 0xfa,
		// 	0x01, 0x54, 0x33, 0x9e, 0xe3, 0x82, 0x9f, 0x30, 0xf0, 0xa4, 0x8f, 0x76, 0xdf, 0x77, 0x11, 0x7e,
		// 	0x56, 0x9e, 0xf3, 0x95, 0xe8, 0xe2, 0x13, 0xb3, 0x1e, 0xb6, 0x70, 0xec, 0x5a, 0x8a, 0xf2, 0x6a,
		// 	0xfc, 0xbc, 0x89, 0x31, 0xe6, 0x7e, 0xe8, 0xb9, 0xc5, 0xf2, 0xc7, 0x1d, 0x78, 0xf3, 0xef, 0x8d,
		// 	0x61, 0xf7, 0x3b, 0xcc, 0x17, 0xc3, 0x40, 0x23, 0x52, 0x4a, 0x8b, 0x9c, 0xb1, 0x75, 0x05, 0x66,
		// 	0xe6, 0xb3}
	} else if data[6] == 3 {
		//these bytes
		log.Println("fairplay type 3")
		// return []byte{0x46, 0x50, 0x4c, 0x59, 0x03, 0x01, 0x04, 0x00, 0x00, 0x00, 0x00, 0x14, 0xaa,
		// 	0x53, 0x69, 0xd9, 0x8d, 0xd5, 0x20, 0xd3, 0xe1, 0x14, 0x59, 0x31, 0xae, 0x83, 0x03, 0xfa, 0x42,
		// 	0x1b, 0xbe, 0xab}

		collect := []byte{0x46, 0x50, 0x4c, 0x59, 0x03, 0x01, 0x04, 0x00, 0x00, 0x00, 0x00, 0x14}
		//[]byte{0x46, 0x50, 0x4c, 0x59, 0x02, 0x01, 0x04, 0x00, 0x00, 0x00, 0x00, 0x14}
		//plus the last 20 bytes of the data
		l := len(data)
		last := data[l-20 : l]
		return append(collect, last...)
	} else {
		log.Println("some other kind of FP setup:", data[6])
	}
	return nil
}

//process the options requests
func (s *airServer) handleOptions(resource string, headers map[string][]string, resHeaders map[string]string) ([]byte, bool) {
	challenge := getHeaderValue(headers, "Apple-Challenge")
	if challenge != "" {
		log.Println("got an Apple-Challenge, need to implement the response")
		//data, err := base64.StdEncoding.DecodeString(challenge)
		//if err == nil {
		//create the challenge response here
		//resHeaders["Apple-Response"] = data
		//}
	}
	resHeaders["Public"] = "ANNOUNCE, SETUP, RECORD, PAUSE, FLUSH, TEARDOWN, OPTIONS, GET_PARAMETER, SET_PARAMETER, POST, GET"
	return nil, true
}

//process announce requests
func (s *airServer) handleAnnounce(con *conn, resource string, headers map[string][]string, data []byte) bool {
	host := s.getClientIP(con)
	c := Client{RTSPUrl: resource, deviceIP: host, Name: getHeaderValue(headers, "X-Apple-Client-Name"), deviceID: getHeaderValue(headers, "X-Apple-Device-ID")}
	//grab the cypto keys from the body
	bodyStr := string(data)
	flags := strings.Split(bodyStr, "\r\n")
	for _, str := range flags {

		if strings.HasPrefix(str, "a=fmtp:") {

		} else if strings.HasPrefix(str, "a=rsaaeskey:") {
			key := str[12:]
			data, err := base64.StdEncoding.DecodeString(key)
			if err != nil {
				log.Println("error decoding RSA key:", err)
				return false
			}
			c.rasAesKey = data
		} else if strings.HasPrefix(str, "a=aesiv:") {
			key := str[8:]
			data, err := base64.StdEncoding.DecodeString(key)
			if err != nil {
				log.Println("error decoding AES key:", err)
				return false
			}
			c.rasAesKey = data
		}
	}
	s.clients[c.deviceIP] = &c
	return true
}

//process the setup requests
func (s *airServer) handleSetup(con *conn, resource string, headers map[string][]string, resHeaders map[string]string) ([]byte, bool) {
	host := s.getClientIP(con)
	c := s.clients[host]
	if c != nil {
		transport := getHeaderValue(headers, "Transport")
		settings := strings.Split(transport, ";")
		for _, setting := range settings {
			a := strings.Split(setting, "=")
			if len(a) > 1 {
				name := a[0]
				if name == "timing_port" {
					port, err := strconv.Atoi(a[1])
					if err != nil {
						log.Println("error on Atoi: ", err)
						return nil, false
					}

					serverPort, controlPort, timePort := c.setup(port, func(b []byte, size int) {
						s.delegate.ReceivedAudioPacket(c, b, size)
					})
					resHeaders["Transport"] = fmt.Sprintf("RTP/AVP/UDP;unicast;mode=record;server_port=%d;control_port=%d;timing_port=%d", serverPort, controlPort, timePort)
					resHeaders["Session"] = "1"
					resHeaders["Audio-Jack-Status"] = "connected"
					return nil, true
				}
			}
		}
	}
	log.Println("failed to do setup")
	return nil, false
}

//process the RECORD requests
func (s *airServer) handleRecord(con *conn, resource string, headers map[string][]string, resHeaders map[string]string) ([]byte, bool) {
	host := s.getClientIP(con)
	c := s.clients[host]
	if c != nil {
		c.start()
		// notify the interface
	}
	return nil, true
}

//process the set_parameters requests
func (s *airServer) handleSetParameters(con *conn, resource string, headers map[string][]string, resHeaders map[string]string) ([]byte, bool) {
	host := s.getClientIP(con)
	c := s.clients[host]
	if c != nil {
		//notify the interface of stuff
	}
	return nil, true
}

//process the FLUSH requests
func (s *airServer) handleFlush(con *conn, resource string, headers map[string][]string, resHeaders map[string]string) ([]byte, bool) {
	host := s.getClientIP(con)
	c := s.clients[host]
	if c != nil {
		c.stop()
		// notify the interface
		delete(s.clients, host)
	}
	return nil, true
}

//process the TEARDOWN requests
func (s *airServer) handleTeardown(con *conn, resource string, headers map[string][]string, resHeaders map[string]string) ([]byte, bool) {
	host := s.getClientIP(con)
	c := s.clients[host]
	if c != nil {
		delete(s.clients, resource)
		c.teardown()
		//notify the interface
	}
	return nil, true
}
