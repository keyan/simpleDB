# simpleDB

A persistent key-value store implementation from the paper ["A simple and efficient implementation of a small database"][paper].

Data is stored in a Go `map` and a write-ahead-log is used to ensure atomic updates. The server can support concurrent operations with key-level locking, so concurrent reads/writes are not permitted. Restarts/crashes will use any available logs and checkpoints stored in `database/data` to reload state during startup. The structure of this directory closely matches the implementation guidelines in the end of Section 3 of the paper.

The client/server communicatation uses [`gob`][gob] over HTTP.

## Usage

Start the DB server with:
```
make server
```

The server interface is limited to:
```
put(key,value)
get(key)
delete(key)
```

Commands can be issued using the client CLI:
```
make client
```

## Automated Run

You can also run an automated simulation which forks and runs a separate server process and issues random client load. A very simple form of process supervision of the server allows for random crashing. The intention of this is to test fault-tolerance and also ensure race conditions are not possible with many concurrent operations. The run takes about ~2min and can be started with:
```
make simulation
```

[paper]: https://dl.acm.org/doi/10.1145/41457.37517
[gob]: https://golang.org/pkg/encoding/gob
