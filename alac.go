///////////////////////////////////////////////////////////////////////////////
// ALAC (Apple Lossless Audio Codec) decoder.
// Ported from David Hammerton C version 2005.
// http://crazney.net/programs/itunes/alac.html
// Also see Apple's OSS version at http://alac.macosforge.org
// Austin Cherry, 2014.
///////////////////////////////////////////////////////////////////////////////

package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

const (
	riceThreshold = 8
)

type AlacFile struct {
	sampleSize     int
	numChannels    int
	bytesPerSample int
	// buffers...
	predicterrorBufferA      []int32
	predicterrorBufferB      []int32
	outputsamplesBufferA     []int32
	outputsamplesBufferB     []int32
	uncompressedBytesBufferA []int32
	uncompressedBytesBufferB []int32
	// stuff from setinfo
	Ignore                    [24]byte
	SetInfoMaxSamplesPerFrame uint32 /* 0x1000 = 4096 */ /* max samples per frame? */
	SetInfo7a                 uint8  /* 0x00 */
	SetInfoSampleSize         uint8  /* 0x10 */
	SetInfoRiceHistorymult    uint8  /* 0x28 */
	SetInfoRiceInitialHistory uint8  /* 0x0a */
	SetInfoRiceKmodifier      uint8  /* 0x0e */
	SetInfo7f                 uint8  /* 0x02 */
	SetInfo80                 uint16 /* 0x00ff */
	SetInfo82                 uint32 /* 0x000020e7 */ /* max sample size?? */
	SetInfo86                 uint32 /* 0x00069fe4 */ /* bit rate (avarge)?? */
	SetInfo8aRate             uint32 /* 0x0000ac44 */
}

// Init a AlacFile struct for us.
func createAlacFile(sampleSize, numberOfChannels int) *AlacFile {
	f := AlacFile{}
	f.sampleSize = sampleSize
	f.numChannels = numberOfChannels
	f.bytesPerSample = (sampleSize / 8) * numberOfChannels
	return &f
}

// decodes an input buffer into an output buffer? Not sure if we need outputSize.
func (f *AlacFile) decodeFrame(inputBuffer, outputBuffer []byte, outputSize int) {

}

// Allocates buffers
func (f *AlacFile) allocateBuffers() {
	f.predicterrorBufferA = make([]int32, f.SetInfoMaxSamplesPerFrame*4)
	f.predicterrorBufferB = make([]int32, f.SetInfoMaxSamplesPerFrame*4)
	f.outputsamplesBufferA = make([]int32, f.SetInfoMaxSamplesPerFrame*4)
	f.outputsamplesBufferB = make([]int32, f.SetInfoMaxSamplesPerFrame*4)
	f.uncompressedBytesBufferA = make([]int32, f.SetInfoMaxSamplesPerFrame*4)
	f.uncompressedBytesBufferB = make([]int32, f.SetInfoMaxSamplesPerFrame*4)
}

// sets the info fields from the byte buffer.
func (f *AlacFile) setInfo(inputBuffer []byte) {
	buffer := bytes.NewReader(inputBuffer)
	if err := binary.Read(buffer, binary.LittleEndian, f); err != nil {
		log.Fatal(err) // probably change this to return an error instead log.
	}

	// swap our little endian value to big endian
	f.SetInfoMaxSamplesPerFrame = swapToBigEndian32(f.SetInfoMaxSamplesPerFrame)
	f.SetInfo80 = swapToBigEndian16(f.SetInfo80)
	f.SetInfo82 = swapToBigEndian32(f.SetInfo82)
	f.SetInfo86 = swapToBigEndian32(f.SetInfo86)
	f.SetInfo8aRate = swapToBigEndian32(f.SetInfo8aRate)

	f.allocateBuffers()
}

// There is probably a faster way to swap these with bit shifting.
func swapToBigEndian32(v uint32) uint32 {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return binary.BigEndian.Uint32(b)
}

// There is probably a faster way to swap these with bit shifting.
func swapToBigEndian16(v uint16) uint16 {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	return binary.BigEndian.Uint16(b)
}

///////////////////////////////////////////////////////////////////////////////
// Stream reading functions.
// Assuming encoding/binary gets us what we need (I doubt it)
// we might need these functions, or might change them from the C version.
///////////////////////////////////////////////////////////////////////////////

// // Supports reading 1 to 16s in big endian format.
// // NOTE. THIS CODE WON'T COMPILE!
// func (f *alacFile) readBits16(bits int) uint32 {
// 	var result uint32
// 	var newAccumulator int

// 	result = (f.inputBuffer[0] << 16) | (f.inputBuffer[1] << 8) | (f.inputBuffer[2])
// 	// shift left by the number of bits we've already read, so that the top 'n' bits of the 24 bits we read will
// 	// be the return bits.
// 	result = result << f.inputBufferBitaccumulator
// 	result = result & 0x00ffffff
// 	// and then only want the top 'n' bits from that, where n is 'bits'.
// 	result = result >> (24 - bits)

// 	newAccumulator = (f.inputBufferBitaccumulator + bits)
// 	// increase the buffer pointer if we've read over n bytes.
// 	f.inputBuffer += (newAccumulator >> 3)

// 	f.inputBufferBitaccumulator = (newAccumulator & 7)

// 	return result
// }

// // supports reading 1 to 32 bits, in big endian format.
// // NOTE. THIS CODE WON'T COMPILE!
// func (f *alacFile) readBits(bits int) uint32 {
// 	var result uint32
// 	if bits > 16 {
// 		bits -= 16
// 		result = f.readBits16(16) << bits
// 	}

// 	result |= f.readBits16(bits)
// 	return result
// }
