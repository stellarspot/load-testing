package snet

import "github.com/DATA-DOG/godog"

func FeatureContext(s *godog.Suite) {

	// run etcd server
	s.Step(`^ectd server is run$`, ectdServerIsRun)

	// run load tests
	s.Step(`^etcd enpoint is "([^"]*)"$`, etcdEnpointIs)
	s.Step(`^there are (\d+) clients$`, thereAreClients)
	s.Step(`^number of iterations is (\d+)$`, numberOfIterationsIs)
	s.Step(`^Put\/Get requests should succeed$`, putGetRequestsShouldSucceed)
	s.Step(`^CompareAndSet requests should succeed$`, compareAndSetRequestsShouldSucceed)
}
