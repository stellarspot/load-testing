package snet

import "github.com/DATA-DOG/godog"

func FeatureContext(s *godog.Suite) {
	s.Step(`^etcd instances names are: (.*)$`, etcdInstancesNamesAre)
	s.Step(`^etcd client enpoints are: (.*)$`, etcdClientEndpointsAre)
	s.Step(`^etcd peer   enpoints are: (.*)$`, etcdPeerEnpointsAre)
	s.Step(`^ectd cluster is run$`, ectdClusterIsRun)
	s.Step(`^there are (\d+) clients$`, thereAreClients)
	s.Step(`^number of iterations is (\d+)$`, numberOfIterationsIs)
	s.Step(`^Put\/Get requests should succeed$`, putGetRequestsShouldSucceed)
	s.Step(`^CompareAndSet requests should succeed$`, compareAndSetRequestsShouldSucceed)
	s.Step(`^Etcd cluster is stopped$`, etcdClusterIsStopped)
}
