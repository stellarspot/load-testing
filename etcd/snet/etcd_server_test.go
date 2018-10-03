package snet

import (
	"errors"
	"time"

	"go.etcd.io/etcd/embed"
)

var endpoints []string = make([]string, 0)

var etcdServer *embed.Etcd

func ectdServerIsRun() error {

	var err error
	cfg := embed.NewConfig()
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
