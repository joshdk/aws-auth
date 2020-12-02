// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package transformers

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/joshdk/aws-auth/config"
	"github.com/joshdk/aws-auth/mfa"
)

const (
	defaultRoleSessionName = "Temp"
	defaultDuration        = time.Hour
)

type AssumeRoleTransform struct {
	Role *config.Role
}

// Transform takes the input sts.Credentials and the internal config.Role and
// performs an AssumeRole. The sts.Credentials for the assumed role are
// returned.
func (s AssumeRoleTransform) Transform(creds *sts.Credentials) (*sts.Credentials, error) {
	// Pack the input struct with appropriate data. Fields that have a
	// zero-value must be nil (opposed if a pointer to a zero-value).
	input := sts.AssumeRoleInput{
		RoleArn: aws.String(s.Role.RoleARN),
	}

	if value := s.Role.DurationSeconds; value != 0 {
		input.DurationSeconds = aws.Int64(int64(value))
	} else {
		input.DurationSeconds = aws.Int64(int64(defaultDuration.Seconds()))
	}

	if value := s.Role.ExternalID; value != "" {
		input.ExternalId = aws.String(value)
	}

	if value := s.Role.MFASerial; value != "" {
		input.SerialNumber = aws.String(value)
	}

	if value := s.Role.RoleSessionName; value != "" {
		input.RoleSessionName = aws.String(value)
	} else {
		input.RoleSessionName = aws.String(defaultRoleSessionName)
	}

	if s.Role.MFASerial != "" {
		// Prompt the user to enter an MFA code.
		code, err := mfa.Prompt(s.Role.MFASerial, s.Role.MFAMessage, s.Role.YubikeySlot)
		if err != nil {
			return nil, err
		}

		input.TokenCode = aws.String(code)
	}

	// Create a session with the input credentials that will be used in the
	// following API call.
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			aws.StringValue(creds.AccessKeyId),
			aws.StringValue(creds.SecretAccessKey),
			aws.StringValue(creds.SessionToken),
		),
	})
	if err != nil {
		return nil, err
	}

	// Perform the actual API call.
	result, err := sts.New(sess).AssumeRole(&input)
	if err != nil {
		return nil, err
	}

	// Return new credentials for this role!
	return result.Credentials, nil
}

// str is a helper that returns a pointer to the given string value, but a nil
// pointer if that value is empty.
func str(value string) *string {
	if value != "" {
		aws.String(value)
	}
	return nil
}
