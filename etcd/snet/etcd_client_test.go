package snet

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type loadSyncStruct struct {
	mx         sync.Mutex
	wg         sync.WaitGroup
	exceptions []error
}

type loadFunc func(s *loadSyncStruct, request *TestClientRequest, storageClient *EtcdStorageClient)

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
	etcdv3  *clientv3.Client
	counter *requestCounter
}

func NewEtcdStorageClient(message string, endpoints []string) (*EtcdStorageClient, error) {
	etcdv3, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
	})

	if err != nil {
		return nil, err
	}

	etcdStorageClient := EtcdStorageClient{
		etcdv3:  etcdv3,
		counter: &requestCounter{},
	}

	return &etcdStorageClient, nil
}

func (storageClient *EtcdStorageClient) Get(key []byte) ([]byte, error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	response, err := storageClient.etcdv3.Get(ctx, byteArraytoString(key))
	defer cancel()
	storageClient.counter.IncReads()

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
	storageClient.counter.IncWrites()

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

	storageClient.counter.IncCAS()

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

	etcd, err := NewEtcdStorageClient("Put/Get requests", clientEndpoints)

	if err != nil {
		return err
	}

	start := time.Now()

	for _, client := range clients {

		key := client.GetKey()
		value := client.GetValue()

		err = etcd.Put(key, value)

		if err != nil {
			return err
		}

		result, err := etcd.Get(key)

		if err != nil {
			return err
		}

		if !bytes.Equal(value, result) {
			return errors.New("Values are not equal")
		}

		client.IncAmount()
	}

	etcd.counter.Count("Put/Get requests", start)

	return nil
}
func loadPutRequestsShouldSucceed() error {

	return loadTest("Load Put/Get requests", func(
		s *loadSyncStruct, request *TestClientRequest,
		storageClient *EtcdStorageClient,
	) {
		defer s.wg.Done()
		for i := 0; i < iterations; i++ {
			key := request.GetKey()
			value := request.GetValue()
			err := storageClient.Put(key, value)

			if err != nil {
				s.mx.Lock()
				defer s.mx.Unlock()
				s.exceptions = append(s.exceptions, err)
			}
		}
	})
}

func loadTest(title string, f loadFunc) error {

	s := &loadSyncStruct{}

	var storageClients []*EtcdStorageClient
	for i := 0; i < len(clients); i++ {
		etcd, err := NewEtcdStorageClient(title, clientEndpoints)
		if err != nil {
			return err
		}
		storageClients = append(storageClients, etcd)
	}

	s.wg.Add(len(clients))
	start := time.Now()

	for index, client := range clients {
		go f(s, client, storageClients[index])
	}

	s.wg.Wait()

	requestCounter := requestCounter{}
	for _, storageClient := range storageClients {
		requestCounter.Add(storageClient.counter)
	}

	requestCounter.Count(title, start)

	if len(s.exceptions) > 0 {
		for _, err := range s.exceptions {
			fmt.Println("Error during put request", err)
		}
		return errors.New("Errors during put requests")
	}

	return nil
}

func compareAndSetRequestsShouldSucceed() error {

	requestCounter := requestCounter{}
	etcd, etcdError := NewEtcdStorageClient("CAS requests", clientEndpoints)

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

	start := time.Now()

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

	requestCounter.Count("Count CAS requests", start)

	return nil
}
