package qubot

// Config is the conguration of Qubot.
type Config struct {
	Database *DatabaseConfig
	Slack    *SlackConfig
	Redmine  *RedmineConfig
}

// DatabaseConfig is the database configuration.
type DatabaseConfig struct {
	Location string
}

// SlackConfig holds the configuration parameters to access Slack.
type SlackConfig struct {
	Nickname string
	Key      string
}

// RedmineConfig is the configuration of Redmine.
type RedmineConfig struct {
	URL           string
	Key           string
	User          string
	VerifyTLSCert bool
}
