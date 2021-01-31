# Server

## Instructions

Must be run inside the `server` folder

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

## MISC

- Generating new token for a user

Tokens are 32 byte long.

```shell
head -c 32 /dev/urandom | base64
```