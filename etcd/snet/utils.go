package snet

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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

func (counter *requestCounter) Count() {
	fmt.Println(counter.message)
	elapsed := time.Now().Sub(counter.start).Seconds()
	requestsPerTime := float64(counter.totalRequets) / float64(elapsed)
	fmt.Println("\x1b[34;1m",
		"total requests: ", counter.totalRequets, "\n",
		"read  requests: ", counter.readRequests, "\n",
		"write requests: ", counter.writeRequests, "\n",
		"cas   requests: ", counter.casRequests, "\n",
		"\x1b[0m",
	)
	fmt.Println("\x1b[36;1m",
		"elapsed time in seconds: ", elapsed, "\n",
		"requests per seconds: ", strconv.FormatFloat(requestsPerTime, 'f', 2, 64), "\n",
		"\x1b[0m",
	)

}

func split(strs string) []string {

	arr := []string{}
	for _, str := range strings.Split(strs, ",") {
		str = strings.Replace(str, "\"", "", -1)
		str = strings.TrimSpace(str)
		arr = append(arr, str)
	}

	return arr
}
