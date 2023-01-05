# baxs2

- config file syntax

`<service name>: <command>`

```
service1: foo bar baz
service2: fooz barz bazz
...
```

- command is run in a shell, i.e. no expansion of env variables and glob patterns

- logs written to central logs dir, stdout and stderr in the same file

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
