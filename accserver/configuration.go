package accserver

import (
	"os/exec"
	"strings"
)

// Configuration specifies the configuration values for managing an accServer
type Configuration struct {
	InstallationDir string `json:"installationDir"`
	ResultsDir      string `json:"resultsDir"`
	NewResultsDelay int    `json:"newResultsDelay"`
	ExeWrapper      string `json:"exeWrapper"`
	LogPrefiltering bool   `json:"logPrefiltering"`
}

// installationDir returns the InstallationDir with a single slash at the end
func (c *Configuration) installationDir() string {
	return strings.TrimRight(c.InstallationDir, "/") + "/"
}

// executable returns the path of accServer.exe
func (c *Configuration) executable() string {
	return c.installationDir() + "accServer.exe"
}

// ResolveResultsDir returns the directory containing the session result files
func (c *Configuration) ResolveResultsDir() string {
	if c.ResultsDir != "" {
		return c.ResultsDir
	}

	return c.installationDir() + "results"
}

// exeWrapper returns the value of ExeWrapper, or the path of wine if ExeWrapper is empty and wine is installed
func (c *Configuration) exeWrapper() string {
	if c.ExeWrapper == "" {
		if wine, err := exec.LookPath("wine"); err == nil {
			return wine
		}
	}
	return c.ExeWrapper
}
