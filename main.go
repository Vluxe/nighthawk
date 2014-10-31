package main

import (
	"fmt"
	"log"
	"net"
)

//get this party started
func main() {
	hardwareAddr := getMacAddress()
	fmt.Println("address:", hardwareAddr)
	serverName := "goAirplay" //the display name of your server
	log.Println("server name:", serverName)

	go startRAOP(hardwareAddr, serverName)
	startAirplay(hardwareAddr, serverName)
}

//gets the mac address to broadcast the airplay service on
func getMacAddress() net.HardwareAddr {
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
