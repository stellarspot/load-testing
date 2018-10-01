package snet

import (
	"context"
	"fmt"
	"time"

	"github.com/DATA-DOG/godog"
	"go.etcd.io/etcd/clientv3"
)

type Client struct {
	channelID int
	nonce     int
	curAmount int
	maxAmount int
	signature int
}

var endpoints []string = make([]string, 0)
var clients []Client = make([]Client, 0)
var readsNum int
var writesNum int
var timeout = 3 * time.Second

func etcdEnpointIs(endpoint string) error {
	endpoints = append(endpoints, endpoint)
	return nil
}

func thereAreClients(clientsNum int) error {

	for i := 0; i < clientsNum; i++ {
		client := Client{
			channelID: i,
		}
		clients = append(clients, client)
	}

	return nil
}

func eachOfThemWritesTimesAndReadsTimes(writes int, reads int) error {

	for i := 0; i < len(clients); i++ {
		fmt.Println("Client is added: ", clients[i])
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
