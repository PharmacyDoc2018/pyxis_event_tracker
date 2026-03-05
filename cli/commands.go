package cli

type CliCommand struct {
	Name        string
	Description string
	Callback    func(*Config) error
}
