# etcd throughput 

## Put/Get

| Requests | Count |
|----------|-------|
| reads    | 6400  |
| writes   | 6400  |
| total    | 12800 |

### Client and etcd on the same host

Elapsed time: 9.48s

Requests per seconds:  1348.89

### Client and etcd on differents hosts

Elapsed time: 11.77s

Requests per seconds:  1086.88


## Comapre And Set

| Requests | Count |
|----------|-------|
| reads    | 12800  |
| cas      | 6400  |
| total    | 19200 |

### Client and etcd on the same host

Elapsed time: 10.938s

Requests per seconds:  1755.30

### Client and etcd on differents hosts

Elapsed time: 16.16s

Requests per seconds:  1188.03

