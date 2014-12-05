package nighthawk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

var (
	mirrorUDPListener udpListener
)

const (
	AIRTUNES_PACKET       = 0x80
	AIRTUNES_TIMING_QUERY = 0xd2
	TIMESTAMP_EPOCH       = 0x83aa7e80 << 32 //not sure if this is right... //2208988800
	mirroringClientPort   = 7010
	mirroringServerPort   = 7011
)

type timeServer struct {
	listener    udpListener
	queryCount  int
	clockOffset uint64
	startTime   uint64
	latency     uint32
	clientPort  int
	clientIP    string
	running     bool
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
		//t.listener.netLn.WriteTo(t.buildQuery(), *addr)
	})
	return t
}

//creates a mirror listener
func createMirrorTimeServer(clientIP string) timeServer {
	t := timeServer{clientPort: mirroringClientPort, clientIP: clientIP}
	t.listener = sharedMirrorListener()
	go t.listener.start(func(b []byte, size int, addr *net.Addr) {
		//process NTP packets
		log.Println("NTP Mirror packet!")
	})
	return t
}

//creates a udp listener struct
func sharedMirrorListener() udpListener {
	if mirrorUDPListener.port == 0 {
		mirrorUDPListener = udpListener{port: mirroringServerPort}
	}
	return mirrorUDPListener
}

//starts the time server by sending the timing packet on a 3 second interval
func (t *timeServer) start() {
	t.running = true
	t.startTime = t.getCurrentNano()
	t.clockOffset = TIMESTAMP_EPOCH
	for t.running {
		t.sendQuery()
		time.Sleep(3 * time.Second)
	}
}

//stops the time server from running
func (t *timeServer) stop() {
	t.running = false
}

//sends a query the the client
func (t *timeServer) sendQuery() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("[%s]:%d", t.clientIP, t.clientPort))
	cAddr, err := net.DialUDP("udp", t.listener.sAddr, addr)
	if err != nil {
		log.Println("unable to start time server:", err)
		return
	}
	count, err := cAddr.Write(t.buildQuery())
	//count, err := t.listener.netLn.WriteToUDP(t.buildQuery(), addr)

	if err != nil {
		log.Println("unable to write to UDP time: ", err)
	}
	log.Println("wrote time packet count:", count)
}

//sends a timing query packet
func (t *timeServer) buildQuery() []byte {
	packet := timingPacket{ident: AIRTUNES_PACKET, command: AIRTUNES_TIMING_QUERY, fixed: swapToBigEndian16(0x0007), zero: 0, timestamp_1: 0,
		timestamp_2: 0, timestamp_3: swapToBigEndian64(t.getTimeStamp() + t.clockOffset)}
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, packet)
	if err != nil {
		log.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

func (t *timeServer) getCurrentNano() uint64 {
	return uint64(time.Now().Nanosecond())
}

//grabs the timestamp from the system clock
func (t *timeServer) getTimeStamp() uint64 {
	stamp := t.getCurrentNano() - t.startTime
	return stamp
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
