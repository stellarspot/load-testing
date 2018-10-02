Feature: etcd load test

  Scenario: write to and read from values
    Given etcd enpoint is "localhost:2379"
    Given there are 30 clients
    Given number of iterations is 30
    Then Put/Get requests should succeed
    Then CompareAndSet requests should succeed
