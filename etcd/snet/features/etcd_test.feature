Feature: etcd load test

  Scenario: write to and read from values
    Given etcd enpoints are: "http://localhost:2370", "http://localhost:2375", "http://localhost:2380" 
    Given ectd server is run
    Given there are 10 clients
    Given number of iterations is 10
    Then Put/Get requests should succeed
    Then CompareAndSet requests should succeed
    Then Etcd server is closed
