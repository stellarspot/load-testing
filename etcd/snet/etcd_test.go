package snet

import (
	"context"
	"fmt"
	"time"

	"github.com/DATA-DOG/godog"
	"go.etcd.io/etcd/clientv3"
)

type Uint32 struct {
	Hash []byte
}

func NewUint32(h []byte) *Uint32 {
	return &Uint32{Hash: h}
}

func IntToUint32(value int) *Uint32 {
	return NewUint32([]byte{
		byte(value & 0x000000FF),
		byte(value & 0x0000FF00),
		byte(value & 0x00FF0000),
		byte(value & 0xFF000000),
		0x00,
		0x00,
		0x00,
		0x00,
	})
}

func Uint32ToInt(uint32 *Uint32) int {

	var value int
	base := 1

	for i := 0; i < 3; i++ {
		value += int(uint32.Hash[i]) * base
		base <<= 8
	}

	return value
}

type Client struct {
	channelID *Uint32
	nonce     *Uint32
	curAmount *Uint32
	maxAmount *Uint32
	signature *Uint32
}

func (client *Client) ToString() string {
	return fmt.Sprint("[",
		"channel id: ", Uint32ToInt(client.channelID), ", ",
		"nonce: ", Uint32ToInt(client.nonce), ", ",
		"curr amount: ", Uint32ToInt(client.curAmount), ", ",
		"signature: ", Uint32ToInt(client.signature),
		"]",
	)
}

var endpoints []string = make([]string, 0)
var clients []Client = make([]Client, 0)
var readsNum int
var writesNum int
var timeout = 3 * time.Second

func createTestClient(clientNum int, requestNum int) Client {
	return Client{
		channelID: IntToUint32(clientNum),
		nonce:     IntToUint32(0),
		curAmount: IntToUint32(10 + clientNum),
		maxAmount: IntToUint32(50),
		signature: IntToUint32(5 + clientNum<<4 + requestNum),
	}
}

func etcdEnpointIs(endpoint string) error {
	endpoints = append(endpoints, endpoint)
	return nil
}

func thereAreClients(clientsNum int) error {

	for i := 0; i < clientsNum; i++ {
		clients = append(clients, createTestClient(i, 3))
	}

	return nil
}

func eachOfThemWritesTimesAndReadsTimes(writes int, reads int) error {

	for i := 0; i < len(clients); i++ {
		fmt.Println("Client is added: ", clients[i].ToString())
	}

	readsNum = reads
	writesNum = writes

	fmt.Println("writes: ", writesNum, "reads: ", readsNum)

	return nil
}

func allResultsShouldSucceed() error {

	fmt.Println("call etcd client")

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
	})

	if err != nil {
		return err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	putResponse, putErr := cli.Put(ctx, "key3", "value3")
	cancel()

	if putErr != nil {
		return putErr
	}

	fmt.Println("put response: ", putResponse)

	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	getResponse, getErr := cli.Get(ctx, "key3")

	cancel()
	if getErr != nil {
		return getErr
	}

	fmt.Println("get response: ", getResponse)

	return nil
}

func FeatureContext(s *godog.Suite) {
	s.Step(`^etcd enpoint is "([^"]*)"$`, etcdEnpointIs)
	s.Step(`^there are (\d+) clients$`, thereAreClients)
	s.Step(`^each of them writes (\d+) times and reads (\d+) times$`, eachOfThemWritesTimesAndReadsTimes)
	s.Step(`^all results should succeed$`, allResultsShouldSucceed)
}
