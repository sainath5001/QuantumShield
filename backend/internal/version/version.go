// Package version holds build-time and default application metadata.
package version

// Values may be overridden via -ldflags at build time.
var (
	// App is the product name.
	App = "Q-Guard"
	// Version is the application version reported by health/info endpoints.
	Version = "0.2.0-history"
	// RiskModelVersion is the active risk classification model identifier.
	RiskModelVersion = "qguard-risk-v1"
)
