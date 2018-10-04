Feature: etcd load test

  Scenario: write to and read from values
    Given etcd instances names are: "infra0" 
    Given etcd client enpoints are: "http://127.0.0.1:2179" 
    Given etcd peer   enpoints are: "http://127.0.0.1:2180" 
    # Given etcd instances names are: "infra0", "infra1" 
    # Given etcd client enpoints are: "http://127.0.0.1:2179", "http://127.0.0.1:2279" 
    # Given etcd peer   enpoints are: "http://127.0.0.1:2180", "http://127.0.0.1:2280" 
    Given ectd server is run
    Given there are 10 clients
    Given number of iterations is 10
    Then Put/Get requests should succeed
    Then CompareAndSet requests should succeed
    Then Etcd server is closed
