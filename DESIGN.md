# Design

The complete solution consists of a backend server serving a RESTfull 
HTTP+JSON interface and a CLI client that connects to this server. 

It is assumed that there is only 1 user using the backend.

## Backend

The backend can start, show status and stop a job. A job is a Unix process.

The backend makes use of a SQL database.

A job can have the status of `RUNNING`, `KILLED` or `FINISHED`.

The `CWD` of the backend is the home folder of a user 
(which ideally should be a user created just for the backend).

### SQL Database

The SQL database will keep track of started jobs and their results.

Table `jobs`:

| id | external_id | pid | status | stdout | stderr | exit_code | created_at | stopped_at |
| -- | ----------- | --- | ------ | ------ | ------ | --------- | ---------- | ------- |
| int | int or string | int | enum (int) | string | string | int | datetime | datetime
|     | NOT NULL, UNIQUE | NOT NULL, INDEXED | |||| NOT NULL, INDEXED | INDEXED
| internal to DB | ID exposed to the client | Unix process ID | status of job | process stdout | process stderr | process exit code | time zwhen job started | time when job killed or finished

The DB will only contain values for `exit_code`, `stopped_at`, `stdout`, `stderr` 
when the process is stopped or finished normally.


### RESTfull API

- Start job: `POST /api/jobs`

  Example body:
  ```javascript
  {
    "program": "ls", // String, path of the process
    "args": ["-l", "./code"] // [String], arguments to start process with
  }
  ```

  Possible HTTP Responses:

  - 201: When job succesfully created
    ```javascript
    {
      "id": "123", // ID which can be used to query status or stop job, it's an internally generated ID
      "status": "RUNNING",
      "created_at": "2020-01-01T12:01Z", // ISO8601 format
    }
    ```

  - 401: Unauthorized

  - 400: when failed to create job


  The backend spawns a thread/goroutine to create the process using `exec` with the 
  arguments as specified in the request body, and `wait`s on it's termination. 
  The stdout and stderr are overriden with pipes to be able to return them to the client. 
  When a job is created the job status is added to the DB with `RUNNING` status and the ID 
  is returned to the client. 
  When the job is finished normally (with exit 0 or otherwise) the DB is updated 
  with `FINIHSED` status and the `stdout`, `stderr` and `exit_code` results are 
  written to the DB, but when a job is stopped by the user the status is set as `KILLED`.

  There is a danger here that a user can execute a malicious job (like `rm -rf /`) which will be ignored.

- List jobs: `GET /api/jobs`

  Possible HTTP Response:

  - 200:
    ```javascript
    [
      {
        "id": "123",
        "status": "RUNNING" | "STOPPED" | "KILLED",
        "created_at": "2020-01-01T12:01Z", // ISO8601 format
        "stopped_at": "2020-02-01T12:01Z", // present if not RUNNING, ISO8601 format
      },
      ...
    ]
    ```

  - 401: Unauthorized

  This is used by the CLI to display the list of jobs.

  The results are not simply a list of IDs because the show status endpoint also returns
  `stdout` which in some cases can be very large, which can be too heavy for the client for
  simply listing the jobs. 

- Show Status: `GET /api/jobs/:id`

  Possible HTTP Responses:
  
  - 200:
    ```javascript
    // if newly created
    {
      "status": "RUNNING",
      "created_at": "2020-01-01T12:01Z"
    }

    // if stopped or the job finished normally
    {
      "status": "KILLED" | "FINISHED",
      "exit_code": 123,
      "stdout": "...",
      "stderr": "...",
      "created_at": "2020-01-01T12:01Z",
      "stopped_at": "2020-02-01T12:01Z"
    }
    ```

  - 401: Unauthorized

  - 404: when invalid ID


- Stop job: `DELETE /api/jobs/:id`

  Possible HTTP Response:

  - 202: When process was signalled to stop

  - 401: Unauthorized

  - 404: when invalid ID

  - 409: when there might be a race condition

  The backend will send a SIGINT signal to the child to stop. 
  The user must then query using the show status to see when it is actually stopped. 
  The signal is only sent when the job has `RUNNING` status. 
  Given that a job might hang it could be possible to add a param to specify wheter to use `SIGKILL`.

  There is a delay from the backend collecting the child job's PID and the backend writting 
  to the DB that the job is no longer `RUNNING`. 
  In this time window another process might have started which could reuse the waited job's PID, 
  and a user could try to stop the already waited job which will cause the new job which has 
  the recycled PID to stop. 
  To solve this the backend could check if there are more than 1 jobs in the DB with the same 
  PID and status `RUNNING` or to check if there is a process with the PID to be stopped but a 
  PPID different from the backend in which case the signal isn't sent and the backend returns 409.

The backend can return errors like 404, 409, these can have a body describing the error with the format:
```javascript
{
  "message": "Something went wrong" // optional
}
```

## CLI Client

The client is a simple stateless, user-friendly, mapping to the backend.

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
ID | STATUS | CREATED AT | STOPPED AT
1 | RUNNING | 2020-01-01T12:01Z | -
2 | KILLED | 2020-01-01T12:01Z | 2020-01-02T12:01Z
```

- Job details
```shellscript
$ jobs-manager details 1
RUNNING, created: 2020-01-01T12:01Z

$ jobs-manager details 2
KILLED, created: 2020-01-01T12:01Z, stopped: 2020-01-02T12:01Z

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
otherwise the HTTP error codes are translated to a human friendly erorr message (on stderr).

## AUTH

### Authentication

The user is authenticated by the backend using mTLS. 
There is only one user, anyone having the private key can use the API.

### Authorization

The user is restricted in the jobs he can execute.

There can be 2 Levels, where on the first level the user can only execute from a whitelist of 
$PATH programs (like `ls`) and on Level 2 can execute all programs in $PATH or anywhere in the system.

When a user tries to execute a LEVEL 2 program but only has LEVEL 1 access the backend returns the 401 status code.

The level of the user can be stored as a config on the backend.