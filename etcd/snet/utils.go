package snet

import "time"

func byteArraytoString(bytes []byte) string {
	return string(bytes)
}

func stringToByteArray(str string) []byte {
	return []byte(str)
}

func intToByte32(value uint32) []byte {
	return []byte{
		byte(value & 0x000000FF),
		byte(value & 0x0000FF00),
		byte(value & 0x00FF0000),
		byte(value & 0xFF000000),
		0x00,
		0x00,
		0x00,
		0x00,
	}
}

func byteArrayToInt(array []byte) int {

	var value int
	base := 1

	for i := 0; i < len(array); i++ {
		value += int(array[i]) * base
		base <<= 8
	}

	return value
}

type requestCounter struct {
	message       string
	totalRequets  int
	readRequests  int
	writeRequests int
	casRequests   int
	start         time.Time
}

func newRequestCounter(msg string) *requestCounter {
	return &requestCounter{message: msg, start: time.Now()}
}

func (counter *requestCounter) IncReads() {
	counter.readRequests++
	counter.totalRequets++
}

func (counter *requestCounter) IncWrites() {
	counter.writeRequests++
	counter.totalRequets++
}

func (counter *requestCounter) IncCAS() {
	counter.casRequests++
	counter.totalRequets++
}
