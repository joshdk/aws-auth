// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package transformers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/joshdk/aws-auth/config"
	"github.com/joshdk/aws-auth/mfa"
)

type SessionTokenTransform struct {
	Session *config.Session
}

// Transform takes the input sts.Credentials and the internal config.Session
// and performs a GetSessionToken. The sts.Credentials for the session are
// returned.
func (s SessionTokenTransform) Transform(creds *sts.Credentials) (*sts.Credentials, error) {
	// Pack the input struct with appropriate data. Fields that have a
	// zero-value must be nil (opposed if a pointer to a zero-value).
	input := sts.GetSessionTokenInput{}

	if value := s.Session.DurationSeconds; value != 0 {
		input.DurationSeconds = aws.Int64(int64(value))
	} else {
		input.DurationSeconds = aws.Int64(int64(defaultDuration.Seconds()))
	}

	if value := s.Session.MFASerial; value != "" {
		input.SerialNumber = aws.String(value)
	}

	if s.Session.MFASerial != "" {
		// Prompt the user to enter an MFA code.
		code, err := mfa.Prompt(s.Session.MFASerial, s.Session.MFAMessage, s.Session.YubikeySlot)
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
	result, err := sts.New(sess).GetSessionToken(&input)
	if err != nil {
		return nil, err
	}

	// Return new credentials for this session!
	return result.Credentials, nil
}
