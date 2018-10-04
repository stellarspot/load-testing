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

	cfg.LCUrls = []url.URL{*clientURL}
	cfg.ACUrls = []url.URL{*clientURL}
	cfg.LPUrls = []url.URL{*peerURL}
	cfg.APUrls = []url.URL{*peerURL}

	initialCluster := getInitialCluster()
	fmt.Printf("initial cluster: '%v'\n", initialCluster)
	cfg.InitialCluster = initialCluster

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
