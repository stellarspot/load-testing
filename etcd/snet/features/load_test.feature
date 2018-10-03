Feature: etcd load test

  Scenario: write to and read from values
    Given ectd server is run
    Given etcd enpoint is "localhost:2379"
    Given there are 10 clients
    Given number of iterations is 10
    Then Put/Get requests should succeed
    Then CompareAndSet requests should succeed
    Then Etcd server is closed
