package main

import (
	"encoding/hex"
	"fmt"
	"github.com/andrewtj/dnssd"
	"log"
	"net"
)

//starts the ROAP service
func startROAP(hardwareAddr net.HardwareAddr, hostName string) {

	name := fmt.Sprintf("%s@%s", hex.EncodeToString(hardwareAddr), hostName)
	op := dnssd.NewRegisterOp(name, "_raop._tcp", 5000, RegisterROAPCallbackFunc)

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
		log.Printf("Failed to register ROAP service: %s", err)
		return
	}
	log.Println("started ROAP service")
	// later...
	//op.Stop()
}

//helper method for the ROAP service
func RegisterROAPCallbackFunc(op *dnssd.RegisterOp, err error, add bool, name, serviceType, domain string) {
	if err != nil {
		// op is now inactive
		log.Printf("ROAP Service registration failed: %s", err)
		return
	}
	if add {
		log.Printf("ROAP Service registered as “%s“ in %s", name, domain)
	} else {
		log.Printf("ROAP Service “%s” removed from %s", name, domain)
	}
}
