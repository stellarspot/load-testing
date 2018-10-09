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

func (s *loadSyncStruct) addError(err error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.exceptions = append(s.exceptions, err)
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
		fmt.Println("CAS response", response)
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

	return loadTest("Load Put/Get requests",
		func(
			s *loadSyncStruct, request *TestClientRequest,
			storageClient *EtcdStorageClient,
		) {
		},
		func(
			s *loadSyncStruct, request *TestClientRequest,
			storageClient *EtcdStorageClient,
		) {
			defer s.wg.Done()
			for i := 0; i < iterations; i++ {
				key := request.GetKey()
				value := request.GetValue()
				err := storageClient.Put(key, value)

				if err != nil {
					s.addError(err)
					break
				}
			}
		})
}

func loadTest(title string, init loadFunc, load loadFunc) error {

	s := &loadSyncStruct{}
	var storages []*EtcdStorageClient

	for _, client := range clients {
		storage, err := NewEtcdStorageClient(title, clientEndpoints)
		if err != nil {
			return err
		}
		storages = append(storages, storage)

		// call init function
		init(s, client, storage)
	}

	err := checkExceptions(s.exceptions)

	if err != nil {
		return err
	}

	s.wg.Add(len(clients))
	start := time.Now()

	for index, client := range clients {
		go load(s, client, storages[index])
	}

	s.wg.Wait()

	requestCounter := requestCounter{}
	for _, storageClient := range storages {
		requestCounter.Add(storageClient.counter)
	}

	requestCounter.Count(title, start)

	return checkExceptions(s.exceptions)
}

func checkExceptions(exceptions []error) error {

	if len(exceptions) > 0 {
		for _, err := range exceptions {
			fmt.Println("Error: ", err)
		}
		return errors.New("Errors during load requests")
	}
	return nil
}

func compareAndSetRequestsShouldSucceed() error {

	etcd, etcdError := NewEtcdStorageClient("CAS requests", clientEndpoints)

	if etcdError != nil {
		return etcdError
	}

	return loadTest("Load CAS requests",
		func(
			s *loadSyncStruct, request *TestClientRequest,
			storageClient *EtcdStorageClient,
		) {
			key := request.GetKey()
			initialAmount := intToByte32(request.prevAmount)

			err := etcd.Put(key, initialAmount)
			if err != nil {
				s.addError(err)
			}
		},
		func(
			s *loadSyncStruct, request *TestClientRequest,
			storageClient *EtcdStorageClient,
		) {
			defer s.wg.Done()

			for i := 0; i < iterations; i++ {

				key := request.GetKey()
				value := request.GetValue()

				prevValue, err := storageClient.Get(key)

				if err != nil {
					s.addError(err)
					break
				}

				success, err := etcd.CompareAndSet(key, prevValue, value)

				if err == nil && !success {
					err = fmt.Errorf("etcd CAS fails expect: %v, update: %v", prevValue, value)
				}

				if err != nil {
					s.addError(err)
					break
				}

				request.IncAmount()
			}
		})
}
