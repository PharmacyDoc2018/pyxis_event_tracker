package cli

import "github.com/PharmacyDoc2018/pyxis_event_tracker/config"

type CliCommand struct {
	Name        string
	Description string
	Callback    func(*config.Config) error
}
