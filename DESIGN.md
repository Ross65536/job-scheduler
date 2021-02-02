# Design

The complete solution consists of a backend server serving a RESTfull 
HTTP+JSON interface and a CLI client that connects to this server. 

Multiple users can communicate with the backend.

The commands and filepaths specified must be absolute paths or in $PATH.

## Backend

The backend can start, show status and stop a job. A job is a Unix process.

A job can have the status of 
- `RUNNING`: executing
- `FINISHED`: job finished normally
- `STOPPING`: user tried to stop the job
- `STOPPED`: job stopped by user (in practice a best case guess is taken if it was actually the user that stopped the process)
- `KILLED`: job stopped by the system

A job belongs to a user that started it. A user can only view or modify his own jobs.

### State

The backend makes use of global state to store information about jobs and users.
The state will keep track of started jobs and their results.

- Jobs:

```golang
type Job struct {
  Id string, // ID exposed to the client (UUID), NOT EMPTY, UNIQUE
  Command []string, // command name + argv, NOT EMPTY
  Status string, // status of job
  Stdout string, // process stdout 
  Stderr string, // process stderr 
  ExitCode int, // process exit code
  CreatedAt time.Time, // time when job started, NOT EMPTY
  StoppedAt time.Time // time when job is killed or has finished
}
```

The `Id` field is a uuid.

The state will only contain values for `ExitCode`, `StoppedAt`
when the process is stopped or finished normally.

The values for `Stdout`, `Stderr` are updated as the job runs and the bytes are available.

- Users: 

```golang
type User struct {
  Username string,
  Token string, // the API token given to the user to access the API, will be generated using a CSPRNG, stored in hex or base64 format
  // Password string, // not used, would be stored using as hash using BCrypt
  Jobs map[string]Job // Index. list of jobs that belong to the user. Index key is the job ID.
}
```

The `Token` is a CSPRNG-random string unique to each user. Would be 32 bytes.

These will be pre-initialized (hardcoded) and will contain the list of valid users.

- Indexes

```golang
var usersIndex map[string][User] // maps username to user struct
```

### RESTfull API

- Start job: `POST /api/jobs`

  Example body:
  ```javascript
  {
    "command": ["ls", "-l", "./code"] // [String], command and arguments to start process with
  }
  ```

  Possible HTTP Responses:

  - 201: When job has been successfully created
    ```javascript
    {
      "id": "123", // ID which can be used to query status or stop job, it's an internally generated ID
      "status": "RUNNING",
      "created_at": "2020-01-01T12:01Z", // ISO8601 format
      "command": ["ls", "-l", "./code"]
    }
    ```

  - 401: On incorrect HTTP Basic credentials

  - 422: When user supplied malformed/invalid JSON

  - 500: when job failed to create because of server error (e. g. OOM, non-existing program, etc)


  The backend spawns a thread/goroutine to create the process using `exec` with the 
  arguments as specified in the request body, and waits on it's termination. 
  The stdout and stderr are overridden with pipes to be able to return them to the client. 
  When a job is created the job status is added to the state with `RUNNING` status and the ID 
  is returned to the client. 
  When the job is finished normally (with exit 0 or otherwise) the state is updated 
  with `FINISHED` status and the `exit_code` is written to the state, but when a job is stopped by the user the status is set as `STOPPED`.
  While the job is running, the spawned thread waits on the overridden `stdout` pipe, as the bytes arrive they are appended to the state, same for `stderr`, when both pipes are closed the thread will wait on the child's PID for it to finish.
  The job belongs to the user that created it (which can be obtained from the API token), when the job is created it is placed in `User.Jobs` under the correct username for `usersIndex`.

  There is a danger here that a user can execute a malicious job (like `rm -rf /`) which will be ignored.

- List jobs: `GET /api/jobs`

  Possible HTTP Response:

  - 200:
    ```javascript
    [
      {
        "id": "123",
        "status": "RUNNING" | "STOPPED" | "KILLED",
        "command": ["ls", "-l", "./code"],
        "created_at": "2020-01-01T12:01Z", // ISO8601 format
        "stopped_at": "2020-02-01T12:01Z", // present if not RUNNING, ISO8601 format
      },
      ...
    ]
    ```

  - 401: On incorrect HTTP Basic credentials

  This is used by the CLI to display the list of jobs for the current user. The list is taken from `User.Jobs`.

  This endpoint could instead be designed to return a list of of job IDs (and nothing else) 
  which the client would then have to query the backend with each ID using the show status 
  endpoint. Because the `stdout` field in some cases can be very large this is not optimal 
  simply for displaying the list of jobs and some minimal information which is what returned 
  here.

- Show Status: `GET /api/jobs/:id`

  Possible HTTP Responses:
  
  - 200:
    ```javascript
    {
      "status": "RUNNING" | "KILLED" | "FINISHED",
      "exit_code": 123,
      "command": ["ls", "-l", "/"],
      "stdout": "...",
      "stderr": "...",
      "created_at": "2020-01-01T12:01Z",
      "stopped_at": "2020-02-01T12:01Z"
    }
    ```
  
  - 401: On incorrect HTTP Basic credentials

  - 404: when invalid ID

  Show job if it belongs to user

- Stop job: `DELETE /api/jobs/:id`

  Possible HTTP Response:

  - 202: When process was signalled to stop

  - 401: On incorrect HTTP Basic credentials

  - 404: when invalid ID

  Stop job if it belongs to user.
  The backend will send a `SIGTERM` signal to the child to stop it and set the status to `STOPPING` if the job's 
  status is `RUNNING`, if the status is already `STOPPING` the `SIGKILL` signal will be sent instead.
  The user must then query using the show status to see when it is actually stopped. 

The backend can return errors like 404, 409, these can have a body describing the error with the format:
```javascript
{
  "status": 500,
  "message": "Something went wrong" // optional
}
```

- Common responses

- 401: The backend will check that the client has supplied the correct API token. This will be used to identify the user.

- 404 (on incorrect ID): the backend checks that the job ID belongs to the user specified by the token by using the `User.Jobs` keys. 

## CLI Client

The client is a simple, user-friendly, 1-to-1 mapping to the backend API.

Example usage:

- Start job
```shellscript
$ jobs-manager start ls -l /
1 # the ID
```

The arguments are sent as is to the backend.

- Listing jobs
```shellscript
$ jobs-manager list
ID | STATUS | COMMAND | CREATED AT | STOPPED AT
1 | RUNNING | ls -l / | 2020-01-01T12:01Z | -
2 | KILLED | task 123 | 2020-01-01T12:01Z | 2020-01-02T12:01Z
```

- Job details
```shellscript
$ jobs-manager show 1
ls -l /, RUNNING, created: 2020-01-01T12:01Z

$ jobs-manager show 2
task 123, KILLED, created: 2020-01-01T12:01Z, stopped: 2020-01-02T12:01Z

STDOUT:
line 1
line 2

STDERR:
line 1 
line 2
```

- Stopping Job

```shellscript
$ jobs-manager stop 1
OK
```

If the backend returns error messages these can be additionally displayed to the user, 
otherwise the HTTP error codes are translated to a human friendly error message (on stderr).

Flags:

These will probably be hardcoded on the client instead.

- The `-c` flag specifies the server location and the HTTP Basic authn for the user. This could be cached between command executions to simplify the interface. Example: `-c user1:pass1@127.0.0.1:8080`
- The `-cert` flag specifies the filepath of the server's public key. 

## Auth

### Authentication

Users are authenticated using HTTP Basic where the user is the username and password is the API token. The credentials are checked against the `usersIndex` global state to see if there is a user and a matching token: the username is used as key for `usersIndex` and the basic password is compared against the `User.Token` field. If the credentials match the user is authenticated, giving the `User` struct which will be used further by Authz.

Client will authenticate the server using the HTTPS certificate. Client will have hardcoded/stored somewhere the server's SSL public key.

### Authorization

The user can only see and stop jobs created by himself. To do that `User.Jobs` is used as described before: new jobs are stored in this index. To check if a job belongs to a user (given a job ID) the `User.Jobs` keys are searched for a matching job ID. The list of user jobs is simply the values of `User.Jobs`.
