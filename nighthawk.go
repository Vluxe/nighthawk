package nighthawk

type airServer struct {
	clients map[string]*Client //the connected clients. Key names are based on the RTSPUrls
}

//Start the airplay server. This will contain closures or an interface of stuff to deal with (like audio/video streams, volume controls, etc)
func Start(serverName string) {
	s := airServer{clients: make(map[string]*Client)}

	// Start broadcasting avaiable services in DNSSD.
	registerServices(serverName)

	// Start the Remote Audio Protocol Server.
	go s.startRAOPServer()

	// Start the Airplay Server.
	s.startAirplay()
}
