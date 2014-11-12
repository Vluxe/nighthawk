package main

import (
	"fmt"
	"github.com/andrewtj/dnssd"
	"log"
	"net/http"
)

//starts the airplay server
func (s *AirServer) startAirplay(hardwareAddr string) {
	//listener, err := net.Listen("tcp", ":7000")
	// if err != nil {
	// 	log.Printf("Listen failed: %s", err)
	// 	return
	// }
	port := 7000 //listener.Addr().(*net.TCPAddr).Port

	op := dnssd.NewRegisterOp(s.ServerName, "_airplay._tcp", port, s.registerCallbackFunc)

	log.Println("hwID:", hardwareAddr)
	op.SetTXTPair("deviceid", hardwareAddr)
	//mask := 0x0 0000 1100 0000 //screen mirroring only
	mask := 0x00C0
	str := fmt.Sprintf("0x%x", mask)
	//str := "0x39f7"
	log.Println("support features: ", str)
	op.SetTXTPair("features", str)
	op.SetTXTPair("model", "AppleTV2,1")
	err := op.Start()
	if err != nil {
		log.Printf("Failed to register airplay service: %s", err)
		return
	}
	log.Println("started Airplay service")

	go s.startMirroringWebServer(7100) //for screen mirroring

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", r.RemoteAddr)
		log.Println("got airplay connection from: ", r.RemoteAddr)
	})

	http.HandleFunc("/server-info", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", r.RemoteAddr)
		log.Println("server info yeah.")
	})

	http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", r.RemoteAddr)
		log.Println("play!!!")
	})
	//log.Println("http port: ", fmt.Sprintf(":%d", port))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	//http.Serve(listener, nil)
	// // later...
	// op.Stop()
}

//helper method for the airplay service
func (s *AirServer) registerCallbackFunc(op *dnssd.RegisterOp, err error, add bool, name, serviceType, domain string) {
	if err != nil {
		// op is now inactive
		log.Printf("Airplay Service registration failed: %s", err)
		return
	}
	if add {
		log.Printf("Airplay Service registered as “%s“ in %s", name, domain)
	} else {
		log.Printf("Airplay Service “%s” removed from %s", name, domain)
	}
}

func (s *AirServer) startMirroringWebServer(port int) {
	StartServer(port, func(c *conn) {
		log.Println("got a Mirror connection from: ", c.rwc.RemoteAddr())
		//s.readRequest(c.buf.Reader)
		c.buf.Write([]byte("OK"))
	})
}
