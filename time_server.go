package nighthawk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

const (
	AIRTUNES_PACKET       = 0x80
	AIRTUNES_TIMING_QUERY = 0xd2
)

type timeServer struct {
	listener    udpListener
	queryCount  int
	clockOffset uint64
	startTime   uint64
	latency     uint32
	clientPort  int
	clientIP    string
}

type timingPacket struct {
	ident       uint8
	command     uint8
	fixed       uint16
	zero        uint32
	timestamp_1 uint64
	timestamp_2 uint64
	timestamp_3 uint64
}

//creates a udp listener struct
func createTimeServer(clientPort int, clientIP string) timeServer {
	t := timeServer{clientPort: clientPort, clientIP: clientIP}
	t.listener = createUDPListener()
	go t.listener.start(func(b []byte, size int, addr *net.Addr) {
		//process NTP packets
		log.Println("NTP packet!")
		//t.listener.netLn.WriteTo(t.sendQuery(), *addr)
	})
	return t
}

//starts the time server by sending the first timing packet
func (t *timeServer) start() {
	cAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("[%s]:%d", t.clientIP, t.clientPort))
	if err != nil {
		log.Println("unable to start time server:", err)
		return
	}
	t.listener.netLn.WriteTo(t.sendQuery(), cAddr)
}

//sends a timing query packet
func (t *timeServer) sendQuery() []byte {
	log.Println("Shooting time packets..")
	packet := timingPacket{ident: AIRTUNES_PACKET, command: AIRTUNES_TIMING_QUERY, fixed: swapToBigEndian16(0x0007), zero: 0, timestamp_1: 0,
		timestamp_2: 0, timestamp_3: swapToBigEndian64(getTimeStamp() + t.clockOffset)}
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, packet)
	if err != nil {
		log.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

//grabs the timestamp from the system clock
func getTimeStamp() uint64 {
	//implement me!!
	return 0
}

//int swap: htonl
func swapToBigEndian32(v uint32) uint32 {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return binary.BigEndian.Uint32(b)
}

//short swap: htons
func swapToBigEndian16(v uint16) uint16 {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	return binary.BigEndian.Uint16(b)
}

//int swap: super huge
func swapToBigEndian64(v uint64) uint64 {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return binary.BigEndian.Uint64(b)
}
