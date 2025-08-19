package version

// Esses valores são sobrescritos via -ldflags no build.
var (
	Service   = "auth-service"
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unknown"
)
