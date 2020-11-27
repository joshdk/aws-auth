package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/ini.v1"
)

const (
	EnvVarAWSConfigFile            = "AWS_CONFIG_FILE"
	EnvVarAWSSharedCredentialsFile = "AWS_SHARED_CREDENTIALS_FILE"
)

type Config struct {
	config      *ini.File
	credentials *ini.File
}

func Load() (*Config, error) {
	// Determine the location of the AWS config file.
	configFile := filepath.Join(userHomeDir(), ".aws", "config")
	if path := os.Getenv(EnvVarAWSConfigFile); path != "" {
		configFile = path
	}

	// Determine the location of the AWS credentials file.
	credentialsFile := filepath.Join(userHomeDir(), ".aws", "credentials")
	if path := os.Getenv(EnvVarAWSSharedCredentialsFile); path != "" {
		credentialsFile = path
	}

	// Parse both the config and credentials file. Since both files are
	// optional, ignore errors if either fails to load/parse.
	var cfg Config
	if file, err := ini.Load(configFile); err == nil {
		cfg.config = file
	}
	if file, err := ini.Load(credentialsFile); err == nil {
		cfg.credentials = file
	}

	// Replace nil files with non-nil, empty files for ease of use.
	if cfg.config == nil {
		cfg.config = ini.Empty()
	}
	if cfg.credentials == nil {
		cfg.credentials = ini.Empty()
	}

	// Delete the "[DEFAULT]" section, only if it's empty, to keep things tidy.
	if len(cfg.config.Section(ini.DefaultSection).Keys()) == 0 {
		cfg.config.DeleteSection(ini.DefaultSection)
	}
	if len(cfg.credentials.Section(ini.DefaultSection).Keys()) == 0 {
		cfg.credentials.DeleteSection(ini.DefaultSection)
	}

	// If both files are completely empty, then error as there's nothing to do.
	if len(cfg.config.Sections()) == 0 && len(cfg.credentials.Sections()) == 0 {
		return nil, fmt.Errorf("configuration files failed to load or were empty")
	}

	return &cfg, nil
}

// userHomeDir returns the home directory for the user the process is
// running under.
func userHomeDir() string {
	if runtime.GOOS == "windows" { // Windows
		return os.Getenv("USERPROFILE")
	}

	// *nix
	return os.Getenv("HOME")
}
