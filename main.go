package main

import (
	"log"
)

//get this party started
func main() {
	serverName := "goAirplay" //the display name of your server
	log.Println("server name:", serverName)
	s := AirServer{ServerName: serverName}
	s.Start()
}
