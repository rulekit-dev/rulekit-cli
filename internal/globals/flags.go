package globals

// Persistent CLI flags shared across all command sub-packages.
// cmd/root.go binds these to cobra persistent flags; cmd/stack and
// cmd/ruleset read them directly.
var (
	Registry  string
	Workspace string
	Dir       string
	Token     string
	Verbose   bool

	// LockfilePath is the path to rulekit.lock, overridable in tests.
	LockfilePath = "rulekit.lock"
)
