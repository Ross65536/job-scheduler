# Client

## Instructions

The shell commands are assuming you are in the `client` folder.

- Running:
```shell
go run main.go
```

- Building:
```shell
# install
go install .
# run, the installed folder might be different
~/go/bin/client
```

- Testing
```shell
go test ./...
```

## Usage

The backend must be running in order to use the client.

The connection flag must be specified for commands other than `help`, example list command:
```shell
$ go run main.go list -c=http://user2:oAtCvE6Xcu07f2PmjoOjq8kv6X2XTgh3s37suKzKHLo\=@localhost:10000
```

The connection flag used in examples will be defined as for example `CON=-c=http://user2:oAtCvE6Xcu07f2PmjoOjq8kv6X2XTgh3s37suKzKHLo\=@localhost:10000` (this user is present in the backend)

> `client` can be the compiled binary or `go run main.go`

Examples:
- Help
  ```shell
  $ client help
  Format: client <command> [-c=<connection>] [id/command]
  ...
  ```

- List All Job
  ```shell
  $ client list $CON
  ID | STATUS | COMMAND | CREATED_AT
  f104e7f5-4601-4e0e-a474-e17e1c375255 | FINISHED | /scripts/long_process.sh 30 | 2021-02-03 20:16:12.52405 +0100 CET
  ...
  ```

- Start Job
  ```shell
  $ client start "ls -l /" $CON
  ID: dc53a7f4-2dc4-42db-863a-de3d788ddff1
  ```

- Show Job Details
  ```shell
  $ client show dc53a7f4-2dc4-42db-863a-de3d788ddff1 $CON
  ls -l /, FINISHED, 2021-02-03 21:41:02.406174 +0100 CET -> 2021-02-03 21:41:02.413084 +0100 CET, exit_code: 0

  STDOUT:
  total 9
  ...

  STDERR:
  ...
  ```

- Stop Job
  ```shell
  $ client stop dc53a7f4-2dc4-42db-863a-de3d788ddff1 $CON
  Stopping
  ```
