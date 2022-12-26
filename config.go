package main

type Config struct {
	Daemon   DaemonConfig
	Services []ServiceConfig
}

type DaemonConfig struct {
	LogsDir string
}

type ServiceConfig struct {
	Name    string
	Command string
}
