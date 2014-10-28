package main

import (
	"fmt"
	"github.com/andrewtj/dnssd"
	"log"
	"net"
	"net/http"
)

//starts the airplay server
func startAirplay(hardwareAddr net.HardwareAddr, name string) {
	listener, err := net.Listen("tcp", ":7000")
	if err != nil {
		log.Printf("Listen failed: %s", err)
		return
	}
	port := listener.Addr().(*net.TCPAddr).Port

	op := dnssd.NewRegisterOp(name, "_airplay._tcp", port, RegisterCallbackFunc)

	log.Println("hwID:", hardwareAddr.String())
	op.SetTXTPair("deviceid", hardwareAddr.String())
	op.SetTXTPair("features", fmt.Sprintf("0x%x", 0x7))
	op.SetTXTPair("model", "AppleTV2,1")
	err = op.Start()
	if err != nil {
		log.Printf("Failed to register airplay service: %s", err)
		return
	}
	log.Println("started airplay service")

	go http.ListenAndServe(":7100", nil)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", r.RemoteAddr)
		log.Println("got connection from: ", r.RemoteAddr)
	})
	http.Serve(listener, nil)
	// // later...
	// op.Stop()
}

//helper method for the airplay service
func RegisterCallbackFunc(op *dnssd.RegisterOp, err error, add bool, name, serviceType, domain string) {
	if err != nil {
		// op is now inactive
		log.Printf("Service registration failed: %s", err)
		return
	}
	if add {
		log.Printf("Service registered as “%s“ in %s", name, domain)
	} else {
		log.Printf("Service “%s” removed from %s", name, domain)
	}
}
