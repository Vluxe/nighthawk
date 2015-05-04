package main

import (
	"fmt"
	"github.com/Lightspeed-Systems/nighthawk"
	//"github.com/davecheney/profile"
	"github.com/nareix/codec"
	"image/jpeg"
	"os"
)

type airplayHandler struct {
}

// Get this party started.
func main() {
	//defer profile.Start(profile.CPUProfile).Stop()
	h := airplayHandler{}
	nighthawk.Start("nighthawk", &h) // set the display name of your server.
}

func (h *airplayHandler) ReceivedAudioPacket(c *nighthawk.Client, data []byte) {
	//log.Println("got an audio packet!")
}

func (h *airplayHandler) ReceivedMirroringPacket(c *nighthawk.Client, data []byte) {
	if c != nil && c.Codec != nil {
		dec, err := codec.NewH264Decoder(c.Codec)
		img, err := dec.Decode(data)
		if err != nil {
			fmt.Println("unabled to create image.", err)
		}
		if err == nil {
			fp, _ := os.Create(fmt.Sprintf("/tmp/dec-%d.jpg", 1))
			jpeg.Encode(fp, img, nil)
			fp.Close()
		}
	} else {
		if c == nil {
			fmt.Println("client is null")
		} else {
			fmt.Println("codec is null")
		}
	}
}

func (h *airplayHandler) SupportedMirrorFeatures() nighthawk.MirrorFeatures {
	return nighthawk.MirrorFeatures{Height: 1280, Width: 720, Overscanned: true, RefreshRate: 0.016666666666666666}
}
