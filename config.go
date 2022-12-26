package main

// Config specifies the configuration parameters needed by the daemon at startup
type Config struct {
	// Daemon configuration
	Daemon DaemonConfig

	// List of services
	Services []ServiceConfig
}

// DaemonConfig specifies the configuration parameters of the daemon
type DaemonConfig struct {
	// Path where to save the logs of the daemon and the services
	LogsDir string
}

// ServiceConfig specifies the configuration parameters of a service
type ServiceConfig struct {
	// Service name, must be unique
	Name string

	// Command to execute
	Command string
}
