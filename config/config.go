// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
	Policy          string
	PolicyARNs      []string
	RoleARN         string
	RoleSessionName string
	SourceProfile   string
	YubikeySlot     string
}

type Session struct {
	DurationSeconds int
	MFAMessage      string
	MFASerial       string
	SourceProfile   string
	YubikeySlot     string
}

type Federate struct {
	DurationSeconds int
	SourceProfile   string
	Name            string
	Policy          string
	PolicyARNs      []string
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
func sectionAsRole(section *ini.Section) (*Role, error) {
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

	// Verify that required fields are present.
	switch {
	case role.RoleARN == "":
		return nil, nil
	case role.SourceProfile == "":
		return nil, nil
	}

	// Use the given duration, or fall back to a 1 hour default.
	if duration, err := section.Key("duration_seconds").Int(); err == nil {
		role.DurationSeconds = duration
	} else {
		role.DurationSeconds = 3600 // 1 hour
	}

	// Read, parse, and combine the referenced policies.
	policyARNs, policy, err := loadPolicies(section.Key("policies").Strings(",")...)
	if err != nil {
		return nil, err
	}

	role.PolicyARNs = policyARNs
	role.Policy = policy
	return &role, nil
}

// sectionAsFederate takes the given ini.Section and converts it to a Federate
// if all of the required fields are present.
func sectionAsFederate(section *ini.Section) (*Federate, error) {
	if !section.Key("federate").MustBool() {
		return nil, nil
	}

	// Pack section values into struct.
	federate := Federate{
		Name:          section.Key("name").Value(),
		SourceProfile: section.Key("source_profile").Value(),
	}

	// Verify that required fields are present.
	switch {
	case federate.SourceProfile == "":
		return nil, nil
	}

	// Use the given duration, or fall back to a 1 hour default.
	if duration, err := section.Key("duration_seconds").Int(); err == nil {
		federate.DurationSeconds = duration
	} else {
		federate.DurationSeconds = 3600 // 1 hour
	}

	// Read, parse, and combine the referenced policies.
	policyARNs, policy, err := loadPolicies(section.Key("policies").Strings(",")...)
	if err != nil {
		return nil, err
	}

	federate.PolicyARNs = policyARNs
	federate.Policy = policy
	return &federate, nil
}

// sectionAsSession takes the given ini.Section and converts it to a Session if
// all of the required fields are present.
func sectionAsSession(section *ini.Section) *Session {
	// Pack section values into struct.
	session := Session{
		MFAMessage:    section.Key("mfa_message").Value(),
		MFASerial:     section.Key("mfa_serial").Value(),
		SourceProfile: section.Key("source_profile").Value(),
		YubikeySlot:   section.Key("yubikey_slot").Value(),
	}

	// Use the given duration, or fall back to a 1 hour default.
	if duration, err := section.Key("duration_seconds").Int(); err == nil {
		session.DurationSeconds = duration
	} else {
		session.DurationSeconds = 3600 // 1 hour
	}

	// Verify that required fields are present.
	switch {
	case session.SourceProfile == "":
		return nil
	default:
		return &session
	}
}

// Profile finds the named profile, and returns only one of either:
// User - Contains credentials.
// Role - Describes how to derive credentials using assume-role.
// Session - Describes how to derive credentials using get-session-token.
// Federate - Describes how to derive credentials using get-federation-token.
// In the event that the named profile does not exist (or is otherwise
// misconfigured), an error is returned.
func (c *Config) Profile(name string) (*User, *Role, *Session, *Federate, error) {
	section, found := c.profile(name)

	// Section is missing altogether.
	if !found {
		return nil, nil, nil, nil, fmt.Errorf("unknown profile")
	}

	// Section contains a User config.
	if user := sectionAsUser(section); user != nil {
		return user, nil, nil, nil, nil
	}

	// Section contains a Federate config.
	if federate, err := sectionAsFederate(section); err != nil {
		// Federate configuration was somehow invalid.
		return nil, nil, nil, nil, err
	} else if federate != nil {
		return nil, nil, nil, federate, nil
	}

	// Section contains a Role config.
	if role, err := sectionAsRole(section); err != nil {
		// Role configuration was somehow invalid.
		return nil, nil, nil, nil, err
	} else if role != nil {
		return nil, role, nil, nil, nil
	}

	// Section contains a Session config. This check must be done after the
	// check for a Role, as a Role config is also a valid Session config.
	if session := sectionAsSession(section); session != nil {
		return nil, nil, session, nil, nil
	}

	// Section doesn't contain any valid configs.
	return nil, nil, nil, nil, fmt.Errorf("invalid profile")
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

// loadPolicies takes a given list of policy references (either an ARN, or a
// file path) and will return a list of all given ARNs, as well as a single IAM
// policy document that is a combination of all given policy files.
func loadPolicies(policyRefs ...string) ([]string, string, error) {
	var documents [][]byte
	var policyARNs []string
	for _, policyRef := range policyRefs {
		if policyRef == "" {
			continue
		}

		// If a policy ARN was given, add it to the list of ARNs. Otherwise,
		// assume it's a file and read the contents.
		if strings.HasPrefix(policyRef, "arn:aws:iam:") {
			policyARNs = append(policyARNs, policyRef)
		} else {
			document, err := ioutil.ReadFile(policyRef)
			if err != nil {
				return nil, "", err
			}
			documents = append(documents, document)
		}
	}

	// Combine all policy documents into one.
	combined, err := combinePolicies(documents...)
	if err != nil {
		return nil, "", err
	}

	return policyARNs, string(combined), nil
}

// combinePolicies takes a given list of JSON-encoded IAM policy documents, and
// returns a new policy document with all of the individual statements combined
// in order.
func combinePolicies(policies ...[]byte) ([]byte, error) {
	// document is used for opaquely unmarshaling all policy statements.
	type document struct {
		Version    string        `json:"Version"`
		Statements []interface{} `json:"Statement"`
	}

	var combined document
	for _, policy := range policies {
		// Unmarshal each policy document.
		var doc document
		if err := json.Unmarshal(policy, &doc); err != nil {
			return nil, err
		}

		// Add each contained statement to the combined policy.
		combined.Version = doc.Version
		for _, statement := range doc.Statements {
			combined.Statements = append(combined.Statements, statement)
		}
	}

	// Avoid returning a policy with no statements.
	if len(combined.Statements) == 0 {
		return nil, nil
	}

	// Return the JSON-encoded combined policy (non-indented to save bytes).
	return json.Marshal(combined)
}
