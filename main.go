package main

import (
	"encoding/hex"
	"fmt"
	"github.com/andrewtj/dnssd"
	"log"
	"net"
	"net/http"
)

func main() {
	doROAP()
	listener, err := net.Listen("tcp", ":7000")
	if err != nil {
		log.Printf("Listen failed: %s", err)
		return
	}
	port := listener.Addr().(*net.TCPAddr).Port

	var hardwareAddr net.HardwareAddr
	hardwareAddr = []byte{0x48, 0x5d, 0x60, 0x7c, 0xee, 0x22}

	op := dnssd.NewRegisterOp("Dalton", "_airplay._tcp", port, RegisterCallbackFunc)

	// var hardwareAddr net.HardwareAddr
	// interfaces, err := net.Interfaces()
	// if err != nil {
	//  log.Println(err)
	// }

	// for _, inter := range interfaces {
	//  hardwareAddr = inter.HardwareAddr
	//  fmt.Println(inter.HardwareAddr)
	// }

	log.Println("hwID:", hardwareAddr.String())
	op.SetTXTPair("deviceid", hardwareAddr.String())
	op.SetTXTPair("features", fmt.Sprintf("0x%x", 0x7))
	op.SetTXTPair("model", "AppleTV2,1")
	err = op.Start()
	if err != nil {
		log.Printf("Failed to register service: %s", err)
		return
	}

	go http.ListenAndServe(":7100", nil)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", r.RemoteAddr)
		log.Println("got connection from: ", r.RemoteAddr)
	})

	http.HandleFunc("/stream.xml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", r.RemoteAddr)
		log.Println("got connection from: ", r.RemoteAddr)
	})
	http.Serve(listener, nil)
	// // later...
	// op.Stop()
}

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

//doRAOP
func doROAP() {
	//listener, err := net.Listen("tcp", ":5000")
	// if err != nil {
	//  log.Printf("Listen failed: %s", err)
	//  return
	// }
	//port := listener.Addr().(*net.TCPAddr).Port

	var hardwareAddr net.HardwareAddr
	hardwareAddr = []byte{0x48, 0x5d, 0x60, 0x7c, 0xee, 0x22}
	name := fmt.Sprintf("%s@%s", hex.EncodeToString(hardwareAddr), "Dalton")
	log.Println("name:", name)

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
		log.Printf("Failed to register service: %s", err)
		return
	}
	log.Println("started ROAP")
	// later...
	//op.Stop()
}

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
