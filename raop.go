package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/andrewtj/dnssd"
	"io"
	"log"
	"net/textproto"
	"strings"
	"sync"
)

const (
	protocolType = "RTSP/1.0"
	carReturn    = "\r\n"
	//rstAvoidanceDelay = 500 * time.Millisecond
)

//starts the RAOP service
func (s *AirServer) startRAOP(hardwareAddr string) {

	port := 5000
	name := fmt.Sprintf("%s@%s", hardwareAddr, s.ServerName)
	op := dnssd.NewRegisterOp(name, "_raop._tcp", port, s.registerRAOPCallbackFunc)

	op.SetTXTPair("txtvers", "1")
	op.SetTXTPair("ch", "2")
	op.SetTXTPair("cn", "0,1,2,3")
	op.SetTXTPair("et", "0,1")
	op.SetTXTPair("sv", "false")
	op.SetTXTPair("da", "true")
	op.SetTXTPair("sr", "44100")
	op.SetTXTPair("ss", "16")
	op.SetTXTPair("pw", "false")
	op.SetTXTPair("vn", "3")
	op.SetTXTPair("tp", "TCP,UDP")
	op.SetTXTPair("md", "0,1,2")
	op.SetTXTPair("vs", "130.14")
	op.SetTXTPair("sm", "false")
	op.SetTXTPair("ek", "1")
	err := op.Start()
	if err != nil {
		log.Printf("Failed to register RAOP service: %s", err)
		return
	}
	log.Println("started RAOP service")
	s.startRAOPServer(port)
	// later...
	//op.Stop()
}

//helper method for the RAOP service
func (s *AirServer) registerRAOPCallbackFunc(op *dnssd.RegisterOp, err error, add bool, name, serviceType, domain string) {
	if err != nil {
		// op is now inactive
		log.Printf("RAOP Service registration failed: %s", err)
		return
	}
	if add {
		log.Printf("RAOP Service registered as “%s“ in %s", name, domain)
	} else {
		log.Printf("RAOP Service “%s” removed from %s", name, domain)
	}
}

//starts the RTSP server
func (s *AirServer) startRAOPServer(port int) {
	StartServer(port, func(c *conn) {
		log.Println("got a RAOP connection from: ", c.rwc.RemoteAddr())
		for {
			verb, resource, headers, data, err := s.readRequest(c.buf.Reader)
			if err != nil {
				return
			}
			resHeaders := make(map[string]string)
			resHeaders["Server"] = "AirPlay/210.98"
			key := "Cseq"
			if headers[key] != nil {
				resHeaders[key] = headers[key][0]
			}
			resData, status := s.processRequest(verb, resource, headers, resHeaders, data)
			c.buf.Write(s.createResponse(status, resHeaders, resData))
			c.buf.Flush()
			c.resetConn()
			//putBufioReader(c.buf.Reader)
			//putBufioWriter(c.buf.Writer)
			// if tcp, ok := c.rwc.(*net.TCPConn); ok {
			// 	log.Println("close and write")
			// 	tcp.CloseWrite()
			// }
			// time.Sleep(rstAvoidanceDelay)
			//c.rwc.Close()
		}
	})
	log.Println("RAOP server finished...?")
}

//creates a response to send back to the client
func (server *AirServer) createResponse(success bool, headers map[string]string, data []byte) []byte {
	s := protocolType
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
	log.Println("response is (minus data):", s)
	body := []byte(s + carReturn)
	if data != nil {
		body = append(body, data...)
	}
	return body
}

//processes the request by dispatching to the proper method for each response
func (s *AirServer) processRequest(verb, resource string, headers map[string][]string, resHeaders map[string]string, data []byte) ([]byte, bool) {
	log.Println("resource is:", resource)
	log.Println("verb is:", verb)
	if verb == "POST" && resource == "/fp-setup" {
		return s.handleFairPlay(resHeaders, data), true
	} else if verb == "OPTIONS" && resource == "*" {
		//do the auth and such
		s.printRequest(verb, resource, headers, data)
	} else if verb == "ANNOUNCE" {
		//we need to collect keys and such here...
		s.printRequest(verb, resource, headers, data)
		return nil, true
	} else if verb == "SETUP" {
		//we need to setup UDP connections and such here
		s.printRequest(verb, resource, headers, data)
		return nil, true
	}
	//more stuff
	return nil, false
}

//temp method for debug purposes
func (s *AirServer) printRequest(verb, resource string, headers map[string][]string, data []byte) {
	//log.Println("resource is:", resource)
	//log.Println("verb is:", verb)
	log.Println("headers:")
	for key, val := range headers {
		log.Printf("key: %s val: %s", key, val)
	}
	log.Println("body: ", string(data))
}

//process fair play requests
func (s *AirServer) handleFairPlay(headers map[string]string, data []byte) []byte {
	if data[6] == 1 {
		return []byte{0x46, 0x50, 0x4c, 0x59, 0x02, 0x01, 0x02, 0x00, 0x00, 0x00, 0x00, 0x82,
			0x02, 0x02, 0x2f, 0x7b, 0x69, 0xe6, 0xb2, 0x7e, 0xbb, 0xf0, 0x68, 0x5f, 0x98, 0x54, 0x7f, 0x37,
			0xce, 0xcf, 0x87, 0x06, 0x99, 0x6e, 0x7e, 0x6b, 0x0f, 0xb2, 0xfa, 0x71, 0x20, 0x53, 0xe3, 0x94,
			0x83, 0xda, 0x22, 0xc7, 0x83, 0xa0, 0x72, 0x40, 0x4d, 0xdd, 0x41, 0xaa, 0x3d, 0x4c, 0x6e, 0x30,
			0x22, 0x55, 0xaa, 0xa2, 0xda, 0x1e, 0xb4, 0x77, 0x83, 0x8c, 0x79, 0xd5, 0x65, 0x17, 0xc3, 0xfa,
			0x01, 0x54, 0x33, 0x9e, 0xe3, 0x82, 0x9f, 0x30, 0xf0, 0xa4, 0x8f, 0x76, 0xdf, 0x77, 0x11, 0x7e,
			0x56, 0x9e, 0xf3, 0x95, 0xe8, 0xe2, 0x13, 0xb3, 0x1e, 0xb6, 0x70, 0xec, 0x5a, 0x8a, 0xf2, 0x6a,
			0xfc, 0xbc, 0x89, 0x31, 0xe6, 0x7e, 0xe8, 0xb9, 0xc5, 0xf2, 0xc7, 0x1d, 0x78, 0xf3, 0xef, 0x8d,
			0x61, 0xf7, 0x3b, 0xcc, 0x17, 0xc3, 0x40, 0x23, 0x52, 0x4a, 0x8b, 0x9c, 0xb1, 0x75, 0x05, 0x66,
			0xe6, 0xb3}
	} else if data[6] == 3 {
		//these bytes
		collect := []byte{0x46, 0x50, 0x4c, 0x59, 0x02, 0x01, 0x04, 0x00, 0x00, 0x00, 0x00, 0x14}
		//plus the last 20 bytes of the data
		l := len(data)
		last := data[l-20 : l]
		return append(collect, last...)
	} else {
		log.Println("some other kind of FP setup:", data[6])
	}
	return nil
}

//some request handling stuff
var textprotoReaderPool sync.Pool

//create a new reader from the pool
func (s *AirServer) newTextprotoReader(br *bufio.Reader) *textproto.Reader {
	if v := textprotoReaderPool.Get(); v != nil {
		tr := v.(*textproto.Reader)
		tr.R = br
		return tr
	}
	return textproto.NewReader(br)
}

//put our reader in the pool
func (s *AirServer) putTextprotoReader(r *textproto.Reader) {
	r.R = nil
	textprotoReaderPool.Put(r)
}

//reads the request and breaks it up in proper chunks
func (server *AirServer) readRequest(b *bufio.Reader) (v string, r string, h map[string][]string, buf []byte, err error) {

	tp := server.newTextprotoReader(b)

	var s string
	if s, err = tp.ReadLine(); err != nil {
		return "", "", nil, nil, err
	}
	defer func() {
		server.putTextprotoReader(tp)
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()
	verb, resource, err := server.parseFirstLine(s)
	if err != nil {
		log.Println("unable to read RAOP request:", err)
		return "", "", nil, nil, err
	}
	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		log.Println("unable to read RAOP mimeHeaders:", err)
		return "", "", nil, nil, err
	}
	count := b.Buffered()
	buffer, _ := b.Peek(count)

	return verb, resource, headers, buffer, nil
}

//parses and returns the verb and resource of the request
func (s *AirServer) parseFirstLine(line string) (string, string, error) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return "", "", errors.New("Invalid RTSP format")
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], nil
}
