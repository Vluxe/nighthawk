package main

import (
	"github.com/Vluxe/nighthawk"
	//"log"
)

type airplayHandler struct {
}

// Get this party started.
func main() {
	h := airplayHandler{}
	nighthawk.Start("nighthawk", &h) // set the display name of your server.
}

func (h *airplayHandler) ReceivedAudioPacket(c *nighthawk.Client, data []byte, length int) {
	//log.Println("got an audio packet")
}

func (h *airplayHandler) SupportedMirrorFeatures() nighthawk.MirrorFeatures {
	return nighthawk.MirrorFeatures{Height: 1280, Width: 720, Overscanned: true, RefreshRate: 0.016666666666666666}
}
