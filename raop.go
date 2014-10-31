package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/andrewtj/dnssd"
	"io"
	"log"
	"net"
	"net/textproto"
	"strings"
	"sync"
	//"bytes"
)

const (
	protocolType = "RTSP/1.0"
	carReturn    = "\r\n"
)

//starts the ROAP service
func startRAOP(hardwareAddr net.HardwareAddr, hostName string) {

	port := 5000
	name := fmt.Sprintf("%s@%s", hex.EncodeToString(hardwareAddr), hostName)
	op := dnssd.NewRegisterOp(name, "_raop._tcp", port, RegisterRAOPCallbackFunc)

	op.SetTXTPair("txtvers", "1")
	op.SetTXTPair("ch", "2")
	op.SetTXTPair("cn", "0,1")
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
	startRAOPServer(port)
	// later...
	//op.Stop()
}

//helper method for the ROAP service
func RegisterRAOPCallbackFunc(op *dnssd.RegisterOp, err error, add bool, name, serviceType, domain string) {
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
func startRAOPServer(port int) {
	StartServer(port, func(c *conn) {
		log.Println("got a RAOP connection from: ", c.rwc.RemoteAddr())
		verb, resource, headers, data, err := readRequest(c.buf.Reader)
		if err != nil {
			return
		}
		var resHeaders map[string]string
		key := "Cseq"
		resHeaders[key] = headers[key][0]
		resHeaders["Server"] = "AirTunes/130.14"
		resData, status := processRequest(verb, resource, &resHeaders, data)
		c.buf.Write(createResponse(status, resHeaders, resData))
		c.rwc.Close()
	})
}

//creates a response to send back to the client
func createResponse(success bool, headers map[string]string, data []byte) []byte {
	s := protocolType
	if success {
		s += " 200 OK" + carReturn
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
func processRequest(verb, resource string, headers *map[string]string, data []byte) ([]byte, bool) {
	log.Println("resource is:", resource)
	log.Println("verb is:", verb)
	if verb == "POST" && resource == "/fp-setup" {
		//do fairplay stuff
	} else if verb == "OPTIONS" && resource == "*" {
		//do the auth and such
	}
	//more stuff
	return nil, false
}

//some request handling stuff
var textprotoReaderPool sync.Pool

//create a new reader from the pool
func newTextprotoReader(br *bufio.Reader) *textproto.Reader {
	if v := textprotoReaderPool.Get(); v != nil {
		tr := v.(*textproto.Reader)
		tr.R = br
		return tr
	}
	return textproto.NewReader(br)
}

//put our reader in the pool
func putTextprotoReader(r *textproto.Reader) {
	r.R = nil
	textprotoReaderPool.Put(r)
}

//reads the request and breaks it up in proper chunks
func readRequest(b *bufio.Reader) (v string, r string, h map[string][]string, buf []byte, err error) {

	tp := newTextprotoReader(b)

	var s string
	if s, err = tp.ReadLine(); err != nil {
		return "", "", nil, nil, err
	}
	defer func() {
		putTextprotoReader(tp)
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()
	verb, resource, err := parseFirstLine(s)
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
func parseFirstLine(line string) (string, string, error) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return "", "", errors.New("Invalid RTSP format")
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], nil
}
