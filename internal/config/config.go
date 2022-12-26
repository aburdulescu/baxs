package config

// Config specifies the configuration parameters needed by the daemon at startup
type Config struct {
	// Daemon configuration
	Daemon Daemon

	// List of services
	Services []Service
}

// Daemon specifies the configuration parameters of the daemon
type Daemon struct {
	// Path where to save the logs of the daemon and the services
	LogsDir string
}

// Service specifies the configuration parameters of a service
type Service struct {
	// Service name, must be unique
	Name string

	// Command to execute
	Command string
}
