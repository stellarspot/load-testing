package snet

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"go.etcd.io/etcd/embed"
)

var mx sync.Mutex

var names []string
var clientEndpoints []string
var peerEndpoints []string

var etcdServers []*embed.Etcd
var serversWaitGroup sync.WaitGroup

func etcdInstancesNamesAre(values string) error {
	names = split(values)
	return nil
}

func etcdClientEndpointsAre(urls string) error {

	clientEndpoints = split(urls)
	return nil
}

func etcdPeerEnpointsAre(urls string) error {

	peerEndpoints = split(urls)
	return nil
}

func ectdClusterIsRun() error {

	serversWaitGroup.Add(len(names))
	var results sync.Map

	for index, name := range names {
		go runServer(name, clientEndpoints[index], peerEndpoints[index], results)
	}

	serversWaitGroup.Wait()

	foundError := false

	results.Range(func(k, v interface{}) bool {
		foundError = true
		fmt.Printf("server: %v, error: %v\n", k, v)
		return true
	})

	if foundError {
		return errors.New("Some servers fail to start. See output for more details")
	}

	return nil
}

func etcdClusterIsStopped() error {

	for _, etcdServer := range etcdServers {
		etcdServer.Server.Stop()
	}
	return nil
}

func getInitialCluster() string {

	initialCluster := ""

	for index, name := range names {
		if index > 0 {
			initialCluster += ","
		}
		initialCluster += name + "=" + peerEndpoints[index]
	}
	return initialCluster
}

func runServer(name string, clientEndpoint string, peerEndPoint string, results sync.Map) {

	defer serversWaitGroup.Done()

	clientURL, err := url.Parse(clientEndpoint)

	if err != nil {
		results.Store(name, err)
		return
	}

	peerURL, err := clientURL.Parse(peerEndPoint)

	if err != nil {
		results.Store(name, err)
		return
	}

	cfg := embed.NewConfig()
	cfg.Name = name
	cfg.Dir = name + ".etcd"

	// --listen-client-urls
	cfg.LCUrls = []url.URL{*clientURL}

	// --advertise-client-urls
	cfg.ACUrls = []url.URL{*clientURL}

	// --listen-peer-urls

	cfg.LPUrls = []url.URL{*peerURL}

	// --initial-advertise-peer-urls
	cfg.APUrls = []url.URL{*peerURL}

	// --initial-cluster
	initialCluster := getInitialCluster()
	fmt.Printf("initial cluster: '%v'\n", initialCluster)
	cfg.InitialCluster = initialCluster

	//  --initial-cluster-state
	cfg.ClusterState = embed.ClusterStateFlagNew

	etcdServer, err := embed.StartEtcd(cfg)

	if err != nil {
		results.Store(name, err)
		return
	}

	mx.Lock()
	etcdServers = append(etcdServers, etcdServer)
	mx.Unlock()

	select {
	case <-etcdServer.Server.ReadyNotify():
	case <-time.After(10 * time.Second):
		etcdServer.Server.Stop()
		results.Store(name, errors.New("etcd server took too long to start"))
		return
	}
}
