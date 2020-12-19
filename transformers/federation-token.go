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
)

type FederationTokenTransform struct {
	Federate *config.Federate
}

const defaultFederationName = "Federated"

// Transform takes the input sts.Credentials and the internal config.Federate
// and performs a GetFederationToken. The sts.Credentials for the federated
// session are returned.
func (s FederationTokenTransform) Transform(creds *sts.Credentials) (*sts.Credentials, error) {
	// Pack the input struct with appropriate data. Fields that have a
	// zero-value must be nil (opposed if a pointer to a zero-value).
	input := sts.GetFederationTokenInput{}

	if value := s.Federate.DurationSeconds; value != 0 {
		input.DurationSeconds = aws.Int64(int64(value))
	} else {
		input.DurationSeconds = aws.Int64(int64(defaultDuration.Seconds()))
	}

	if value := s.Federate.Policy; value != "" {
		input.Policy = aws.String(value)
	}

	for _, value := range s.Federate.PolicyARNs {
		input.PolicyArns = append(input.PolicyArns, &sts.PolicyDescriptorType{
			Arn: aws.String(value),
		})
	}

	if value := s.Federate.Name; value != "" {
		input.Name = aws.String(value)
	} else {
		input.Name = aws.String(defaultFederationName)
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
	result, err := sts.New(sess).GetFederationToken(&input)
	if err != nil {
		return nil, err
	}

	// Return new credentials for this session!
	return result.Credentials, nil
}
