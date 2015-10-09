package config

import (
	"fmt"

	"qubot"

	multierror "github.com/hashicorp/go-multierror"
)

// Validate the confifugration file
func Validate(c *qubot.Config) error {
	var result error

	if c.Database == nil {
		result = multierror.Append(result, fmt.Errorf("'database' configuration section is missing"))
	}
	if c.Slack == nil {
		result = multierror.Append(result, fmt.Errorf("'slack' configuration section is missing"))
	}
	if c.Redmine == nil {
		result = multierror.Append(result, fmt.Errorf("'redmine' configuration section is missing"))
	}

	if c.Slack != nil {
		if c.Slack.Key == "" {
			result = multierror.Append(result, fmt.Errorf("slack: key is required"))
		}
	}

	if c.Redmine != nil {
		if c.Redmine.URL == "" {
			result = multierror.Append(result, fmt.Errorf("redmine: url is required"))
		}
		if c.Redmine.Key == "" {
			result = multierror.Append(result, fmt.Errorf("redmine: key is required"))
		}
	}

	return result
}
