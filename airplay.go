package nighthawk

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	//"encoding/pem"
	"github.com/DHowett/go-plist"
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

type streamPlist struct {
	DeviceID  int    `plist:"deviceID"`
	SessionID int    `plist:"sessionID"`
	Version   string `plist:"version"`
	Param1    []byte `plist:"param1"`
	Param2    []byte `plist:"param2"`
	LatencyMS int    `plist:"latencyMs"`
}

const (
	airplayPort      = 7000
	mirroringPort    = 7100
	HTTPProtocolType = "HTTP/1.1"
	//rstAvoidanceDelay = 500 * time.Millisecond
)

var pemString, _ = base64.StdEncoding.DecodeString(`MIIEpQIBAAKCAQEA59dE8qLieItsH1WgjrcFRKj6eUWqi+bGLOX1HL3U3GhC/j0Qg90u3sG/1CUt
wC5vOYvfDmFI6oSFXi5ELabWJmT2dKHzBJKa3k9ok+8t9ucRqMd6DZHJ2YCCLlDRKSKv6kDqnw4U
wPdpOMXziC/AMj3Z/lUVX1G7WSHCAWKf1zNS1eLvqr+boEjXuBOitnZ/bDzPHrTOZz0Dew0uowxf
/+sG+NCK3eQJVxqcaJ/vEHKIVd2M+5qL71yJQ+87X6oV3eaYvt3zWZYD6z5vYTcrtij2VZ9Zmni/
UAaHqn9JdsBWLUEpVviYnhimNVvYFZeCXg/IdTQ+x4IRdiXNv5hEewIDAQABAoIBAQDl8Axy9XfW
BLmkzkEiqoSwF0PsmVrPzH9KsnwLGH+QZlvjWd8SWYGN7u1507HvhF5N3drJoVU3O14nDY4TFQAa
LlJ9VM35AApXaLyY1ERrN7u9ALKd2LUwYhM7Km539O4yUFYikE2nIPscEsA5ltpxOgUGCY7b7ez5
NtD6nL1ZKauw7aNXmVAvmJTcuPxWmoktF3gDJKK2wxZuNGcJE0uFQEG4Z3BrWP7yoNuSK3dii2jm
lpPHr0O/KnPQtzI3eguhe0TwUem/eYSdyzMyVx/YpwkzwtYL3sR5k0o9rKQLtvLzfAqdBxBurciz
aaA/L0HIgAmOit1GJA2saMxTVPNhAoGBAPfgv1oeZxgxmotiCcMXFEQEWflzhWYTsXrhUIuz5jFu
a39GLS99ZEErhLdrwj8rDDViRVJ5skOp9zFvlYAHs0xh92ji1E7V/ysnKBfsMrPkk5KSKPrnjndM
oPdevWnVkgJ5jxFuNgxkOLMuG9i53B4yMvDTCRiIPMQ++N2iLDaRAoGBAO9v//mU8eVkQaoANf0Z
oMjW8CN4xwWA2cSEIHkd9AfFkftuv8oyLDCG3ZAf0vrhrrtkrfa7ef+AUb69DNggq4mHQAYBp7L+
k5DKzJrKuO0r+R0YbY9pZD1+/g9dVt91d6LQNepUE/yY2PP5CNoFmjedpLHMOPFdVgqDzDFxU8hL
AoGBANDrr7xAJbqBjHVwIzQ4To9pb4BNeqDndk5Qe7fT3+/H1njGaC0/rXE0Qb7q5ySgnsCb3DvA
cJyRM9SJ7OKlGt0FMSdJD5KG0XPIpAVNwgpXXH5MDJg09KHeh0kXo+QA6viFBi21y340NonnEfdf
54PX4ZGS/Xac1UK+pLkBB+zRAoGAf0AY3H3qKS2lMEI4bzEFoHeK3G895pDaK3TFBVmD7fV0Zhov
17fegFPMwOII8MisYm9ZfT2Z0s5Ro3s5rkt+nvLAdfC/PYPKzTLalpGSwomSNYJcB9HNMlmhkGzc
1JnLYT4iyUyx6pcZBmCd8bD0iwY/FzcgNDaUmbX9+XDvRA0CgYEAkE7pIPlE71qvfJQgoA9em0gI
LAuE4Pu13aKiJnfft7hIjbK+5kyb3TysZvoyDnb3HOKvInK7vXbKuU4ISgxB2bB3HcYzQMGsz1qJ
2gG0N5hvJpzwwhbhXqFKA4zaaSrw622wDniAK5MlIE0tIAKKP4yxNGjoD2QYjhBGuhvkWKY=`)

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
				if !s.handleVideoStream(c) {
					break
				}
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
					host := s.getClientIP(c)
					client := s.clients[host]
					if client != nil {
						var bplist streamPlist
						_, err := plist.Unmarshal(data, &bplist)
						if err != nil {
							fmt.Println(err)
						}
						// client.aesKey = make([]byte, base64.StdEncoding.DecodedLen(len(bplist.Param1)))
						// client.aesIV = make([]byte, base64.StdEncoding.DecodedLen(len(bplist.Param2)))
						// count, err := base64.StdEncoding.Decode(client.aesKey, bplist.Param1)
						// if err != nil {
						// 	log.Println("base64 Param1 decode error:", err)
						// } else {
						// 	log.Println("base64 Param1 decode count:", count)
						// }
						// count2, err2 := base64.StdEncoding.Decode(client.aesIV, bplist.Param2)
						// if err2 != nil {
						// 	log.Println("base64 Param2 decode error:", err2)
						// } else {
						// 	log.Println("base64 Param2 decode count:", count2)
						// }
						log.Println("latencyMs:", bplist.LatencyMS)
						log.Println("deviceID:", bplist.DeviceID)
						log.Println("param1:", string(bplist.Param1))
						log.Println("param2:", string(bplist.Param2))
						client.aesKey = s.decryptKey(bplist.Param1)
						client.aesIV = bplist.Param2
						client.startMirror()
					} else {
						log.Println("airplay client is nil")
					}
					isStream = true //set if the stream has change from HTTP to video
				}
			}
			c.buf.Flush()
		}
		c.rwc.Close()
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

// handleVideoStream "decodes" the video packets.
func (s *airServer) handleVideoStream(c *conn) bool {
	buffer := make([]byte, 128)
	_, err := c.Read(buffer)
	if err != nil { //&& err != io.EOF
		log.Println(err)
		return false
	}
	var header streamPacketHeader
	buf := bytes.NewReader(buffer)
	err = binary.Read(buf, binary.LittleEndian, &header)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	switch header.PayloadType {
	case 0:
		//fmt.Println("bitstream.")
		s.handleVideoBitStream(c, &header)
	case 1:
		//fmt.Println("codec.")
		s.handleCodecData(c, &header)
	case 2:
		//fmt.Println("heartbeat.")
	}
	return true
}

// handleVideoBitStream handles the actual video packets and sends them off to interface.
func (s *airServer) handleVideoBitStream(c *conn, header *streamPacketHeader) {
	if header.PayloadSize > 0 {
		buffer := s.readPayload(c, header)
		host := s.getClientIP(c)
		client := s.clients[host]
		if client != nil {
			if client.aesKey != nil {
				buf := s.decryptPacket(client, buffer)          // might want this to update buffer instead of return a new one.
				s.delegate.ReceivedMirroringPacket(client, buf) // send the video packet off to be handled.
			} else {
				s.delegate.ReceivedMirroringPacket(client, buffer) // send the video packet off to be handled.
			}
		}
	}
}

func (s *airServer) handleCodecData(c *conn, header *streamPacketHeader) {
	if header.PayloadSize > 0 {
		buffer := s.readPayload(c, header)
		host := s.getClientIP(c)
		client := s.clients[host] // get our client.
		if client != nil {
			client.Codec = buffer // save the codec info to our client.
		} else {
			log.Println("not a valid client on: ", c.rwc.RemoteAddr())
		}
	}
}

// readPayload reads the payload of video packet. Not sure if airServer is the right place.
func (s *airServer) readPayload(c *conn, header *streamPacketHeader) []byte {
	buffer := make([]byte, header.PayloadSize)
	n, err := c.Read(buffer)
	if err != nil {
		log.Println(err)
	}
	for int32(n) < header.PayloadSize { // make sure we read all the bytes.
		count, err := c.Read(buffer)
		if err != nil {
			log.Println(err)
			break
		}
		if count <= 0 {
			break
		}
		n += count
	}
	return buffer
}

//decrypts the RSA key so it can be used to decrypt the packets
func (s *airServer) decryptKey(param1 []byte) []byte {
	log.Println("param1 FP type: ", param1[6])
	privKey, err := x509.ParsePKCS1PrivateKey(pemString) //p.Bytes
	if err != nil {
		fmt.Println(err)
	}

	sha1 := sha1.New()
	random := rand.Reader

	key, err := rsa.DecryptOAEP(sha1, random, privKey, param1, nil)
	if err != nil {
		fmt.Println("D-nice", err)
		return param1
	}
	return key
}

// decryptPacket does AES CBC decryption on video packets.
func (s *airServer) decryptPacket(client *Client, buffer []byte) []byte {
	// p, _ := pem.Decode(pemString)
	// if p == nil {
	// 	fmt.Println("You done messed up A-aron.")
	// }
	// rsakey, err := x509.ParsePKCS1PrivateKey(pemString) //p.Bytes
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// random := rand.Reader
	// sha1 := sha1.New()
	// key, err := rsa.DecryptOAEP(sha1, random, rsakey, client.aesKey, nil)
	// if err != nil {
	// 	fmt.Println("D-nice", err)
	// 	return buffer
	// }

	block, err := aes.NewCipher(client.aesKey) //key
	if err != nil {
		fmt.Println("cipher error", err)
		return buffer
	}
	iv := client.aesIV
	mode := cipher.NewCBCDecrypter(block, iv)
	dst := make([]byte, len(buffer))
	mode.CryptBlocks(dst, buffer)
	return dst
}
