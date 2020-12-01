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

type User struct {
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSSessionToken    string
}

type Role struct {
	DurationSeconds int
	ExternalID      string
	MFAMessage      string
	MFASerial       string
	RoleARN         string
	RoleSessionName string
	SourceProfile   string
	YubikeySlot     string
}

// sectionAsUser takes the given ini.Section and converts it to a User if all
// of the required fields are present.
func sectionAsUser(section *ini.Section) *User {
	// Pack section values into struct.
	// https://docs.aws.amazon.com/cli/latest/topic/config-vars.html#credentials
	user := User{
		AWSAccessKeyID:     section.Key("aws_access_key_id").Value(),
		AWSSecretAccessKey: section.Key("aws_secret_access_key").Value(),
		AWSSessionToken:    section.Key("aws_session_token").Value(),
	}

	// Verify that required fields are present.
	switch {
	case user.AWSAccessKeyID == "":
		return nil
	case user.AWSSecretAccessKey == "":
		return nil
	default:
		return &user
	}
}

// sectionAsRole takes the given ini.Section and converts it to a Role if all
// of the required fields are present.
func sectionAsRole(section *ini.Section) *Role {
	// Pack section values into struct.
	// https://docs.aws.amazon.com/cli/latest/topic/config-vars.html#using-aws-iam-roles
	role := Role{
		ExternalID:      section.Key("external_id").Value(),
		MFAMessage:      section.Key("mfa_message").Value(),
		MFASerial:       section.Key("mfa_serial").Value(),
		RoleARN:         section.Key("role_arn").Value(),
		RoleSessionName: section.Key("role_session_name").Value(),
		SourceProfile:   section.Key("source_profile").Value(),
		YubikeySlot:     section.Key("yubikey_slot").Value(),
	}

	// Use the given duration, or fall back to a 1 hour default.
	if duration, err := section.Key("duration_seconds").Int(); err == nil {
		role.DurationSeconds = duration
	} else {
		role.DurationSeconds = 3600 // 1 hour
	}

	// Verify that required fields are present.
	switch {
	case role.RoleARN == "":
		return nil
	case role.SourceProfile == "":
		return nil
	default:
		return &role
	}
}

// Profile finds the named profile, and returns only one of either a User
// (which contains credentials), or a Role (which describes how to derive
// credentials). In the event that the named profile does not exist (or is
// otherwise misconfigured), both return values will be nil.
func (c *Config) Profile(name string) (*User, *Role) {
	section, found := c.profile(name)

	// Section is missing altogether.
	if !found {
		return nil, nil
	}

	// Section contains valid User settings.
	if user := sectionAsUser(section); user != nil {
		return user, nil
	}

	// Section contains valid Role settings.
	if role := sectionAsRole(section); role != nil {
		return nil, role
	}

	// Section doesn't contain any valid settings.
	return nil, nil
}

// profile looks up the given section name from the AWS config/credentials
// file, following the rules for section naming and precedence in those files.
func (c *Config) profile(name string) (*ini.Section, bool) {
	// Look up the profile name directly in the credentials file.
	if section, err := c.credentials.GetSection(name); err == nil {
		return section, true
	}

	// In the config file, profile names are prefixed with "profile ", except
	// for the "default" profile which is used verbatim.
	// "default" → "default"
	// "example" → "profile example"
	if name != "default" {
		name = "profile " + name
	}

	// Look up the (possibly prefixed) profile name in the config file.
	if section, err := c.config.GetSection(name); err == nil {
		return section, true
	}

	return nil, false
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
