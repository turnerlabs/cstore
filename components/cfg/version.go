package cfg

// Version is overridden by the main function unless tests
// which require a default are running.
var Version = "v4.0.0-default"

// SupportedVersions ...
var SupportedVersions = []string{"v2", "v3", Version}
