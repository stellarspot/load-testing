package snet

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.etcd.io/etcd/embed"
)

var endpoints []string

var etcdServer *embed.Etcd

func etcdEnpointIs(endpoint string) error {
	endpoints = append(endpoints, endpoint)
	return nil
}

func ectdServerIsRun() error {

	var err error
	var url *url.URL

	endpoint := endpoints[0]
	fmt.Println("endpoint: ", endpoint)

	url, err = url.Parse(endpoint)

	if err != nil {
		return err
	}

	cfg := embed.NewConfig()
	cfg.LCUrls = append(cfg.LCUrls, *url)
	cfg.Dir = "default.etcd"

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

func etcdServerIsClosed() error {
	etcdServer.Server.Stop()
	return nil
}
