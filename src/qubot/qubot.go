package qubot

import "config"

// Qubot represents the bot service.
type Qubot struct {
	config *config.Config
	db     *DB
}

// New returns a Qubot server
func New(config *config.Config) (*Qubot, error) {
	db := DB{}
	if err := db.Open(config.Database.Location, 0640); err != nil {
		return nil, err
	}
	defer db.Close()

	q := Qubot{
		config: config,
		db:     &db,
	}

	return &q, nil
}

// Reload the configuration of Qubit
func Reload() chan error {
	return nil
}
