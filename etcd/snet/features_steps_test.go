package snet

import "github.com/DATA-DOG/godog"

func FeatureContext(s *godog.Suite) {
	s.Step(`^etcd enpoints are: (.*)$`, etcdEnpointsAre)
	s.Step(`^ectd server is run$`, ectdServerIsRun)
	s.Step(`^there are (\d+) clients$`, thereAreClients)
	s.Step(`^number of iterations is (\d+)$`, numberOfIterationsIs)
	s.Step(`^Put\/Get requests should succeed$`, putGetRequestsShouldSucceed)
	s.Step(`^CompareAndSet requests should succeed$`, compareAndSetRequestsShouldSucceed)
	s.Step(`^Etcd server is closed$`, etcdServerIsClosed)
}
