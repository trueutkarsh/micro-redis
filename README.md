# Micro Redis
This project is an aim to create a redis like in-memory database server and its client that support a subset of operations mentioned below
- GET
- DEL
- EXPIRE
- KEYS
- SET
- TTL
ONLY for String datatype

Here are some references used for this project
https://redis.io
https://redis.io/docs/reference/protocol-spec
https://redis.io/commands/set


## How to run

### Server
```bash
go run cmd/server/main.go -address={address} -port={port} -clearfreq={clearfreq}
```
where ```address``` and ```port``` are the address port you want
to run the server from. The default values for these flags are
localhost and 6379. The third flag is clearfreq which determines at
what rate should the expired keys be cleared out of storage in milliseconds

### Client
```bash
go run cmd/client/main.go -address={address} -port={port}
```
where ```address``` and ```port``` are the address port you want
to run the client from and connect the server to.



**Note**
'#' is now a special keyword and can't be used inside keys and values currently. Doing so will result in undefined behaviour. This is a known bug and will be resolved.