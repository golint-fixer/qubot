package qubot

// PluginSpec holds the specification of a plugin that may be registered with
// Qubot.
type PluginSpec struct {
	Name  string
	Help  string
	Start func(p *Plugger) Stopper
}

// Stopper is implemented by types that can run arbitrary background
// activities that can be stopped on request.
type Stopper interface {
	Stop() error
}

type pluginManager struct {
}

var registeredPlugins = make(map[string]*PluginSpec)

// RegisterPlugin registers with Qubot the plugin defined via the provided
// specification, so that it may be loaded when configured to be.
func RegisterPlugin(spec *PluginSpec) {
	if spec.Name == "" {
		panic("Cannot register plugin with an empty name")
	}
	if _, ok := registeredPlugins[spec.Name]; ok {
		panic("Plugin already registered: " + spec.Name)
	}
	registeredPlugins[spec.Name] = spec
}

// Plugger provides the interface between a plugin and the bot infrastructure.
type Plugger struct {
}
