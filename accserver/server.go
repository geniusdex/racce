package accserver

import "strings"

// Configuration specifies the configuration values for managing an accServer
type Configuration struct {
	InstallationDir string `json:"installationDir"`
	ResultsDir      string `json:"resultsDir"`
	NewResultsDelay int    `json:"newResultsDelay"`
}

// installationDir returns the InstallationDir with a single slash at the end
func (c *Configuration) installationDir() string {
	return strings.TrimRight(c.InstallationDir, "/") + "/"
}

// ResolveResultsDir returns the directory containing the session result files
func (c *Configuration) ResolveResultsDir() string {
	if c.ResultsDir != "" {
		return c.ResultsDir
	}

	return c.installationDir() + "results"
}
