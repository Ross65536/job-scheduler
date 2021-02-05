# Job Manager

A server and CLI for starting/stopping/getting jobs.

Go `1.15` is recommended.

Consists of a:

- Server
- Client

There are useful scripts in the `scripts` directory for testing the behavior of the system.

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

The connection flag must be specified for commands other than `help`, example list command:
```shell
$ go run src/cmd/client/main.go -c=http://user2:oAtCvE6Xcu07f2PmjoOjq8kv6X2XTgh3s37suKzKHLo\=@localhost:10000 list
```

The connection flag used in examples will be defined as for example `CON=-c=http://user2:oAtCvE6Xcu07f2PmjoOjq8kv6X2XTgh3s37suKzKHLo\=@localhost:10000` (this user is present in the backend)

> `client` can be the compiled binary or `go run src/cmd/client/main.go`

Examples:
- Help
  ```shell
  $ client help
  Format: client <command> [-c=<connection>] [id/command]
  ...
  ```

- List All Jobs
  ```shell
  $ client $CON list 
  ID | STATUS | COMMAND | CREATED_AT
  f104e7f5-4601-4e0e-a474-e17e1c375255 | FINISHED | /scripts/long_process.sh 30 | 2021-02-03 20:16:12.52405 +0100 CET
  ...
  ```

- Start Job
  ```shell
  $ client $CON start "ls -l /" 
  ID: dc53a7f4-2dc4-42db-863a-de3d788ddff1
  ```

- Show Job Details
  ```shell
  $ client $CON show dc53a7f4-2dc4-42db-863a-de3d788ddff1
  ls -l /, FINISHED, 2021-02-03 21:41:02.406174 +0100 CET -> 2021-02-03 21:41:02.413084 +0100 CET, exit_code: 0

  STDOUT:
  total 9
  ...

  STDERR:
  ...
  ```

- Stop Job
  ```shell
  $ client $CON stop dc53a7f4-2dc4-42db-863a-de3d788ddff1
  Stopping
  ```

## Server

### MISC

- Generating new token for a user

Tokens are 32 byte long.

```shell
head -c 32 /dev/urandom | base64
```