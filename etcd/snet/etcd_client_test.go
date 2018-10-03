package snet

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func (counter *requestCounter) Count() {
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
	return intToByte32(client.channelID)
}

func (client *TestClientRequest) GetValue() []byte {
	return intToByte32(client.curAmount)
}

func (client *TestClientRequest) GetPrevValue() []byte {
	return intToByte32(client.prevAmount)
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
	response, err := storageClient.etcdv3.Get(ctx, byteArraytoString(key))
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
	response, err := etcdv3.Put(ctx, byteArraytoString(key), byteArraytoString(value))
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
		clientv3.Compare(clientv3.Value(byteArraytoString(key)), "=", byteArraytoString(expect)),
	).Then(
		clientv3.OpPut(byteArraytoString(key), byteArraytoString(update)),
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

	requestCounter := newRequestCounter("Count Put/Get requests")

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

	requestCounter := newRequestCounter("Count CAS requests")
	etcd, etcdError := NewEtcdStorageClient(endpoints)

	if etcdError != nil {
		return etcdError
	}

	// init
	for _, client := range clients {

		key := client.GetKey()
		initialAmount := intToByte32(client.prevAmount)

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
