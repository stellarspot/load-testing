Feature: etcd load test

  Scenario: write to and read from values
    Given etcd enpoint is "localhost:2379"
    Given there are 5 clients
    Given number of iterations is 10
    Then Put/Get requests should succeed
