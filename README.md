# Job Manager

A server and CLI for starting/stopping/getting jobs.

Go `1.15` is recommended.

Consists of a:

- Server
- Client

There are useful scripts in the `scripts` directory for testing the behavior of the system.

Quick start:
```shell
# clone the repository into your $GOPATH

# must be in the root folder for the defaults
$ pwd 
~/go/src/github.com/Ross65536/job-scheduler

# setup
$ go get ./...

# start server on :10000
$ go run src/cmd/server/main.go 

# on another terminal, send client command
$ go run src/cmd/client/main.go start ls -l /

# process with non zero exit code 
$ go run src/cmd/client/main.go start $(pwd)/scripts/bad_exit.sh 1

# stopping long running process
$ go run src/cmd/client/main.go start $(pwd)/scripts/long_process.sh 30
ID: 218fd3b8-201b-47ef-a3c0-37aba0403263 # ID returned can be different

$ go run src/cmd/client/main.go stop 218fd3b8-201b-47ef-a3c0-37aba0403263 # replace ID with result from previous step
```

## Instructions

You need to clone the repository into your `GOPATH` env var folder (which needs to be set).

If your `GOPATH` is `~/go` you need to clone into `~/go/src/github.com/ros-k/job-manager`

Assuming you are in the root folder:

- Get dependencies:
```shell
go get ./...
```

- Testing: 
```shell
go test ./...
```

- Installing:
```shell
go install ./...
# the binaries will be in '~/go/bin'
~/go/bin/client
~/go/bin/server
```

- Running:
```shell
# client
go run src/cmd/client/main.go
# server
go run src/cmd/server/main.go
```

## Client

The backend must be running in order to use the client.

> `client` can be the compiled binary or `go run src/cmd/client/main.go`

Examples:
- Help
  ```shell
  $ client -h
  Format: client <command> [-c=<connection>] [id/command]
  ...
  ```

- List All Jobs
  ```shell
  $ client list 
  ID | STATUS | COMMAND | CREATED_AT
  f104e7f5-4601-4e0e-a474-e17e1c375255 | FINISHED | /scripts/long_process.sh 30 | 2021-02-03 20:16:12.52405 +0100 CET
  ...
  ```

- Start Job
  ```shell
  $ client start "ls -l /" 
  ID: dc53a7f4-2dc4-42db-863a-de3d788ddff1
  ```

- Show Job Details
  ```shell
  $ client show dc53a7f4-2dc4-42db-863a-de3d788ddff1
  ls -l /, FINISHED, 2021-02-03 21:41:02.406174 +0100 CET -> 2021-02-03 21:41:02.413084 +0100 CET, exit_code: 0

  STDOUT:
  total 9
  ...

  STDERR:
  ...
  ```

- Stop Job
  ```shell
  $ client stop dc53a7f4-2dc4-42db-863a-de3d788ddff1
  Stopping
  ```

### Flags

These flags have defaults for running on localhost from the root folder.

- `c`: The URL connection flag can be specified. A default is set for localhost
- `ca`: Private CA path to the public key. Will be added by the client as a trusted CA, which can be used to sign server certificates.

Example full command:
```shell
$ go run src/cmd/client/main.go -ca=certs/rootCA.crt -c=https://user2:oAtCvE6Xcu07f2PmjoOjq8kv6X2XTgh3s37suKzKHLo\=@localhost:10000 list
```

## Server

### Flags

These flags have defaults for running on localhost from the root folder.

- `p`: the port to listen on
- `cert`: path to the server's public certificate for HTTPS
- `privateKey`: path to the server's private key for HTTPS

Example full command:
```shell
$ go run src/cmd/server/main.go -cert=certs/server.crt -privateKey=certs/server.key
```

### MISC

#### Generating new token for a user

Tokens are 32 byte long.

```shell
head -c 32 /dev/urandom | base64
```

#### Generating certificates

Assuming you are in the `certs` folder.

> In answer to the question `Common Name (eg, fully qualified host name)` below you should set `localhost` (or your real domain)
> Using `secp521r1` elliptic curve. `secp384r1` and `prime256v1` could also be used if `secp521r1` not supported.

- Generate private CA

```shell
openssl ecparam -genkey -name secp521r1 -out rootCA.key
# the CA public key, send to client to have it trust it
openssl req -x509 -new -key rootCA.key -days 3650 -out rootCA.crt
```

- Generate server certificate

```shell
# generate server keys, these must be passed to the server application
openssl ecparam -genkey -name secp521r1 -out server.key
openssl req -new -key server.key -out server.csr

# sign server key with CA
openssl x509 -req -in server.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -days 365 -out server.crt -extensions SAN -extfile <(cat /etc/ssl/openssl.cnf \
    <(printf "\n[SAN]\nsubjectAltName=DNS:localhost")) # here must put server domain
```