package snet

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/DATA-DOG/godog"
	"go.etcd.io/etcd/clientv3"
)

func ByteArraytoString(bytes []byte) string {
	return string(bytes)
}

func StringToByteArray(str string) []byte {
	return []byte(str)
}

func IntToByte32(value uint32) []byte {
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

func ByteArrayToInt(array []byte) int {

	var value int
	base := 1

	for i := 0; i < len(array); i++ {
		value += int(array[i]) * base
		base <<= 8
	}

	return value
}

type RequestCounter struct {
	message       string
	totalRequets  int
	readRequests  int
	writeRequests int
	casRequests   int
	start         time.Time
}

func NewRequestCounter(msg string) *RequestCounter {
	return &RequestCounter{message: msg, start: time.Now()}
}

func (counter *RequestCounter) IncReads() {
	counter.readRequests++
	counter.totalRequets++
}

func (counter *RequestCounter) IncWrites() {
	counter.writeRequests++
	counter.totalRequets++
}

func (counter *RequestCounter) IncCAS() {
	counter.casRequests++
	counter.totalRequets++
}

func (counter *RequestCounter) Count() {
	fmt.Println(counter.message)
	elapsed := time.Now().Sub(counter.start).Seconds()
	requestsPerTime := float64(counter.totalRequets) / float64(elapsed)
	fmt.Println("total requests: ", counter.totalRequets,
		"read  requests: ", counter.readRequests,
		"write requests: ", counter.writeRequests,
		"cas   requests: ", counter.casRequests,
	)
	fmt.Println("elapsed time in seconds: ", elapsed,
		"requests per seconds: ", strconv.FormatFloat(requestsPerTime, 'f', 2, 64))

}

type TestClientRequest struct {
	channelID  uint32
	nonce      uint32
	prevAmount uint32
	curAmount  uint32
	maxAmount  uint32
	signature  uint32
}

func (client *TestClientRequest) GetKey() []byte {
	return IntToByte32(client.channelID)
}

func (client *TestClientRequest) GetValue() []byte {
	return IntToByte32(client.curAmount)
}

func (client *TestClientRequest) GetPrevValue() []byte {
	return IntToByte32(client.prevAmount)
}

func (client *TestClientRequest) IncAmount() {
	client.prevAmount = client.curAmount
	client.curAmount = client.curAmount + 1
}

func (client *TestClientRequest) ToString() string {
	return fmt.Sprint("[",
		"channel id: ", client.channelID, ", ",
		// "nonce: ", client.nonce, ", ",
		"prev amount: ", client.prevAmount, ", ",
		"curr amount: ", client.curAmount, ", ",
		// "signature: ", client.signature,
		"]",
	)
}

type EtcdStorageClient struct {
	etcdv3 *clientv3.Client
}

func NewEtcdStorageClient(endpoints []string) (*EtcdStorageClient, error) {
	etcdv3, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
	})

	if err != nil {
		return nil, err
	}

	etcdStorageClient := EtcdStorageClient{
		etcdv3: etcdv3,
	}

	return &etcdStorageClient, nil
}

func (storageClient *EtcdStorageClient) Get(key []byte) ([]byte, error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	response, err := storageClient.etcdv3.Get(ctx, ByteArraytoString(key))
	defer cancel()

	if debug {
		fmt.Println("get response", response)
	}

	if err != nil {
		return nil, err
	}

	for _, kv := range response.Kvs {
		return kv.Value, nil
	}
	return nil, nil
}

func (storageClient *EtcdStorageClient) Put(key []byte, value []byte) error {

	etcdv3 := storageClient.etcdv3
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	response, err := etcdv3.Put(ctx, ByteArraytoString(key), ByteArraytoString(value))
	defer cancel()

	if debug {
		fmt.Println("put response", response)
	}

	return err
}

func (storageClient *EtcdStorageClient) CompareAndSet(key []byte, expect []byte, update []byte) (bool, error) {

	etcdv3 := storageClient.etcdv3
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	response, err := etcdv3.KV.Txn(ctx).If(
		clientv3.Compare(clientv3.Value(ByteArraytoString(key)), "=", ByteArraytoString(expect)),
	).Then(
		clientv3.OpPut(ByteArraytoString(key), ByteArraytoString(update)),
	).Commit()

	if debug {
		fmt.Println("put response", response)
	}

	if err != nil {
		return false, err
	}

	return response.Succeeded, nil
}

var debug = false
var endpoints []string = make([]string, 0)

var clients []*TestClientRequest
var clientsNum int
var iterations int
var timeout = 3 * time.Second

func createTestClient(clientNum uint32) *TestClientRequest {

	prevAmount := uint32(1)

	return &TestClientRequest{
		channelID:  clientNum,
		nonce:      clientNum * 2,
		prevAmount: prevAmount,
		curAmount:  prevAmount + 1,
		maxAmount:  50,
		signature:  100,
	}
}

func etcdEnpointIs(endpoint string) error {
	endpoints = append(endpoints, endpoint)
	return nil
}

func thereAreClients(num int) error {

	clientsNum = num
	for i := 0; i < clientsNum; i++ {
		clients = append(clients, createTestClient(uint32(i)))
	}

	return nil
}

func numberOfIterationsIs(iter int) error {
	iterations = iter
	fmt.Println("iterations: ", iterations)
	return nil
}

func putGetRequestsShouldSucceed() error {

	requestCounter := NewRequestCounter("Count Put/Get requests")

	etcd, err := NewEtcdStorageClient(endpoints)

	if err != nil {
		return err
	}

	for i := 0; i < iterations; i++ {

		// put/get
		for _, client := range clients {

			key := client.GetKey()
			value := client.GetValue()

			err = etcd.Put(key, value)
			requestCounter.IncWrites()

			if err != nil {
				return err
			}

			result, err := etcd.Get(key)
			requestCounter.IncReads()

			if err != nil {
				return err
			}

			if !bytes.Equal(value, result) {
				return errors.New("Values are not equal")
			}

			client.IncAmount()
		}
	}

	requestCounter.Count()

	return nil
}

func compareAndSetRequestsShouldSucceed() error {

	requestCounter := NewRequestCounter("Count CAS requests")
	etcd, etcdError := NewEtcdStorageClient(endpoints)

	if etcdError != nil {
		return etcdError
	}

	// init
	for _, client := range clients {

		key := client.GetKey()
		initialAmount := IntToByte32(client.prevAmount)

		err := etcd.Put(key, initialAmount)
		if err != nil {
			return err
		}
	}

	for i := 0; i < iterations; i++ {

		// CAS
		for _, client := range clients {

			key := client.GetKey()
			value := client.GetValue()

			prevValue, err := etcd.Get(key)
			requestCounter.IncReads()

			if err != nil {
				return err
			}

			success, err := etcd.CompareAndSet(key, prevValue, value)
			requestCounter.IncCAS()

			if err != nil {
				return err
			}

			newValue, err := etcd.Get(key)
			requestCounter.IncReads()

			if !success {
				msg := fmt.Sprintln("etcd CAS fails. Values",
					" expect: ", prevValue,
					" update: ", value,
					" result: ", newValue,
				)
				return errors.New(msg)
			}

			client.IncAmount()
		}
	}

	requestCounter.Count()

	return nil
}

func FeatureContext(s *godog.Suite) {
	s.Step(`^etcd enpoint is "([^"]*)"$`, etcdEnpointIs)
	s.Step(`^there are (\d+) clients$`, thereAreClients)
	s.Step(`^number of iterations is (\d+)$`, numberOfIterationsIs)
	s.Step(`^Put\/Get requests should succeed$`, putGetRequestsShouldSucceed)
	s.Step(`^CompareAndSet requests should succeed$`, compareAndSetRequestsShouldSucceed)
}
