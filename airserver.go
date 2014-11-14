package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
)

type AirServer struct {
	ServerName string             //the display name of your server that will be broadcast
	Clients    map[string]*Client //the connected clients. Key names are based on the RTSPUrls
}

//Start the airplay server. This will contain closures or an interface of stuff to deal with (like audio/video streams, volume controls, etc)
func (s *AirServer) Start() {
	s.Clients = make(map[string]*Client)
	hardwareAddr := s.getMacAddress()
	fmt.Println("address:", hardwareAddr)
	//start the Remote Audio Protocol, this is DNS and TCP servers
	go s.startRAOP(hex.EncodeToString(hardwareAddr))
	//start the Airplay protocol, this is DNS and TCP servers
	s.startAirplay(hardwareAddr.String())
}

//gets the mac address to broadcast the airplay service on
func (s *AirServer) getMacAddress() net.HardwareAddr {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
	}

	for _, inter := range interfaces {
		if inter.HardwareAddr != nil && len(inter.HardwareAddr) > 0 && inter.Flags&net.FlagLoopback == 0 && inter.Flags&net.FlagUp != 0 && inter.Flags&net.FlagMulticast != 0 && inter.Flags&net.FlagBroadcast != 0 {
			//log.Println("found possible address: ", inter.HardwareAddr)
			addrs, _ := inter.Addrs()
			for _, addr := range addrs {
				if addr.String() != "" {
					//log.Println("found the address: ", inter.HardwareAddr)
					//log.Println("IP: ", addr.String())
					return inter.HardwareAddr
				}
			}
		}
	}
	log.Println("WARNING: didn't find mac address, using default one")
	return []byte{0x48, 0x5d, 0x60, 0x7c, 0xee, 0x22} //default because we couldn't find the real one
}
