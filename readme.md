# baxs

Background service supervisor.

## Config file syntax

Similar to [Procfile](https://devcenter.heroku.com/articles/procfile)
(i.e. `<service name>: <command>`) but with support for line comments:

```
# this is a comment

service1: foo bar baz

service2: fooz barz bazz
```

## Features

- command is not run in a shell => no expansion of env variables and glob patterns

- logs written to central logs dir; stdout and stderr in the same file

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
