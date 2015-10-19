package taunt

import "qubot"

// Plugin holds the plugin specification.
var Plugin = qubot.PluginSpec{
	Name:  "taunt",
	Help:  "Command people to respect Qubot!",
	Start: start,
}

func init() {
	qubot.RegisterPlugin(&Plugin)
}

type tauntPlugin struct {
	plugger *qubot.Plugger
}

func start(plugger *qubot.Plugger) qubot.Stopper {
	p := &tauntPlugin{plugger: plugger}
	return p
}

func (p *tauntPlugin) Stop() error {
	return nil
}

func (p *tauntPlugin) HandleCommand() {}
