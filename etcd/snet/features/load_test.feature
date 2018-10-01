Feature: etcd load test

  Scenario: write to and read from values
    Given etcd enpoint is "localhost:2379"
    Given there are 3 clients
    Given each of them writes 10 times and reads 20 times
    Then all results should succeed
