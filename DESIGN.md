# Design

The complete solution consists of a backend server serving a RESTfull 
HTTP+JSON interface and a CLI client that connects to this server. 

Multiple users can communicate with the backend.

## Backend

The backend can start, show status and stop a job. A job is a Unix process.

A job can have the status of `RUNNING`, `KILLED` or `FINISHED`.

The `CWD` of the backend is the home folder of a user 
(which ideally should be a user created just for the backend).

### 'Database'

The backend makes use of global state to store information about jobs.
The state will keep track of started jobs and their results.

The state is a hash of structs:

| id | pid | command | status | stdout | stderr | exit_code | created_at | stopped_at |
| ----------- | --- | ------- | ------ | ------ | ------ | --------- | ---------- | ------- |
| string (hash key) | int | string |  enum (int) | string | string | int | datetime | datetime
| NOT NULL, UNIQUE | NOT NULL | |||| NOT NULL | 
| ID exposed to the client | Unix process ID | command name + argv | status of job | process stdout | process stderr | process exit code | time when job started | time when job is killed or has finished

The `id` field is a uuid.

The state will only contain values for `exit_code`, `stopped_at`
when the process is stopped or finished normally.

The values for `stdout`, `stderr` are updated as the job runs and the bytes are available.


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

  - 400: when failed to create job

  - 401: On incorrect HTTP Basic credentials


  The backend spawns a thread/goroutine to create the process using `exec` with the 
  arguments as specified in the request body, and waits on it's termination. 
  The stdout and stderr are overridden with pipes to be able to return them to the client. 
  When a job is created the job status is added to the state with `RUNNING` status and the ID 
  is returned to the client. 
  When the job is finished normally (with exit 0 or otherwise) the state is updated 
  with `FINISHED` status and the `exit_code` is written to the state, but when a job is stopped by the user the status is set as `KILLED`.
  While the job is running, the spawned thread waits on the overridden `stdout` pipe, as the bytes arrive they are appended to the state, same for `stderr`, when both pipes are closed the thread will wait on the child's PID for it to finish.

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

  This is used by the CLI to display the list of jobs.

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


- Stop job: `DELETE /api/jobs/:id`

  Possible HTTP Response:

  - 202: When process was signalled to stop

  - 401: On incorrect HTTP Basic credentials

  - 404: when invalid ID

  The backend will send a `SIGTERM` signal to the child to stop. 
  The user must then query using the show status to see when it is actually stopped. 
  The signal is only sent when the job has `RUNNING` status. 
  Given that a job might hang it could be possible to add a param to specify whether to use `SIGKILL`.

The backend can return errors like 404, 409, these can have a body describing the error with the format:
```javascript
{
  "message": "Something went wrong" // optional
}
```

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

Users are authenticated using HTTP Basic.
Client will authenticate the server using the HTTPS certificate. Client will have pinned the public key.

### Authorization

The user is restricted in the jobs he can execute.

There can be 2 Levels, where on the first level the user can only execute from a whitelist of 
$PATH programs (like `ls`) and on Level 2 can execute all programs in $PATH or anywhere in the system.

When a user tries to execute a LEVEL 2 program but only has LEVEL 1 access the backend returns the 401 status code.

The level of the user can be stored as a config on the backend.
