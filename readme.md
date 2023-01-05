# baxs

Background service supervisor.

## Config file syntax(`baxsfile`)

The config file consists of lines of text.

A line can be a comment or a service definition.

A comment starts with a `#`, leading whitespaces(spaces or tabs) are ignored.
What follows after `#` is the body of the comment.

`# this is a comment`

A service definition has the following form:

`<service name>: <command>`

Example:

```
# run a nginx webserver
nginx: /usr/sbin/nginx

# some other service
random service: foo bar baz
```

## Features

- command is not run in a shell => no expansion of env variables and glob patterns

- logs written to central logs dir, one per service(stdout and stderr in the same file)

- daemon will log to stdout and will be controlled via the system process supervisor(e.g. systemd)

- ctl commands will be passed to daemon via unix socket

- ctl commands:

    ```
    ls
    start <cmd_name>
    stop <cmd_name>
    restart <cmd_name>
    ```

    If `cmd_name` is not specified, the ctl command will be applied to all.

    No ctl command for reloading baxfile and starting againg after that, use daemon command for that.

    No ctl commands for logs, they are stored in files in a known folder on disk,
    so use standard shell tools to examine them(cat, grep, tail etc.).

## Usage

- write a `baxsfile`(i.e. baxs config file):

    ```
    sleeper_forever: sleep infinity
    sleeper_3600s: sleep 3600
    ```

- start the daemon, giving it the path where to save the logs and the path to the config file

    ```
    ./baxs daemon -l /tmp/baxs_logs -f /etc/baxsfile
    ```

- check status of services

    ```
    ./baxs ls
    ```

- stop service

    ```
    baxs stop sleeper_forever
    ```

- start service

    ```
    baxs start sleeper_forever
    ```
