// Package registry defines types and operations for managing the agent
// registry, including data types, remote fetch, and local cache.
package registry

// InstallMethod identifies how an agent installation command is executed.
type InstallMethod string

const (
	// MethodCurlBash indicates installation via curl | bash pipeline.
	MethodCurlBash InstallMethod = "curl_bash"

	// MethodNpmInstall indicates installation via npm install -g <package>.
	MethodNpmInstall InstallMethod = "npm_install"

	// MethodCustom indicates a custom installation command.
	MethodCustom InstallMethod = "custom"
)

// InstallCmd describes how to install and optionally uninstall an agent.
type InstallCmd struct {
	Method         InstallMethod `json:"method"`
	URL            string        `json:"url"`
	Command        string        `json:"command"`
	NonInteractive bool          `json:"non_interactive"`
	UninstallCmd   string        `json:"uninstall,omitempty"`
}

// RuntimeDep describes a runtime dependency required by an agent.
type RuntimeDep struct {
	Runtime    string `json:"runtime"`
	MinVersion string `json:"min_version,omitempty"`
}

// Checksum holds verification data for an agent's installation artifact.
type Checksum struct {
	SHA256     string `json:"sha256"`
	VerifiedAt string `json:"verified_at"`
}

// Agent represents a single AI coding agent available in the registry.
type Agent struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Version      string       `json:"version"`
	DetectCmd    string       `json:"detect_command"`
	Install      InstallCmd   `json:"install"`
	Dependencies []RuntimeDep `json:"dependencies"`
	Checksum     *Checksum    `json:"checksum,omitempty"`
	Tags         []string     `json:"tags"`
	AddedAt      string       `json:"added_at"`
}

// Catalog is the top-level container for the agent registry payload.
type Catalog struct {
	Agents    []Agent `json:"agents"`
	UpdatedAt string  `json:"updated_at"`
}
