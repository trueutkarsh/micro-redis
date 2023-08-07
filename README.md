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
```bash
go run main.go
```
This will start a redis server in the background and redis client in your terminal.
Both redis client and server will operate from localhost:6379 address by default.
You can change it manually inside main.go file.


**Note**
'#' is now a special keyword and can't be used inside keys and values currently. Doing so will result in undefined behaviour. This is a known bug and will be resolved.