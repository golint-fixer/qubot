package app

const (
	// Name is the program name.
	Name = "Qubot"

	// Version is the main version number that is being run at the moment.
	Version = "0.1.0"

	// VersionPrerelease is a pre-release marker for the version. If this is ""
	// (empty string) then it means that it is a final release. Otherwise, this is a
	// pre-release such as "dev" (in development), "beta", "rc1", etc.
	VersionPrerelease = "dev"
)

var (
	// Revision is the git commit that was compiled. This will be filled in
	// by the compiler.
	Revision string
)
