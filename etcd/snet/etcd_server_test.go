package snet

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.etcd.io/etcd/embed"
)

var clientEndpoints []string
var peerEndpoints []string
var names []string

var etcdServer *embed.Etcd

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

func ectdServerIsRun() error {

	for index, name := range names {
		runServer(name, clientEndpoints[index], peerEndpoints[index])
	}
	return nil
}

func etcdServerIsClosed() error {
	etcdServer.Server.Stop()
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

func runServer(name string, clientEndpoint string, peerEndPoint string) error {

	clientURL, err := url.Parse(clientEndpoint)

	if err != nil {
		return err
	}

	peerURL, err := clientURL.Parse(peerEndPoint)

	if err != nil {
		return err
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

	// cfg.Debug = true

	etcdServer, err = embed.StartEtcd(cfg)

	if err != nil {
		return err
	}

	select {
	case <-etcdServer.Server.ReadyNotify():
	case <-time.After(10 * time.Second):
		etcdServer.Server.Stop()
		return errors.New("etcd server took too long to start")
	}

	return nil
}
