# Server

## Instructions

The shell commands are assuming you are in the `server` folder.

Server listens on http://localhost:10000

- Running:
```shell

go run main.go
```

- Building:
```shell
# install
go install .
# run, the installed folder might be different
~/go/bin/server
```

- Testing
```shell
go test ./...
```

## MISC

- Generating new token for a user

Tokens are 32 byte long.

```shell
head -c 32 /dev/urandom | base64
```