package main

import (
	"github.com/Vluxe/nighthawk"
	"log"
)

type airplayHandler struct {
}

// Get this party started.
func main() {
	h := airplayHandler{}
	nighthawk.Start("nighthawk", &h) // set the display name of your server.
}

func (h *airplayHandler) ReceivedAudioPacket(c *nighthawk.Client, data []byte, length int) {
	log.Println("got an audio packet")
}
