// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package transformers

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/joshdk/aws-auth/config"
)

// Chain finds a sequence of transforms that can be used to obtain credentials
// for the given profile, starting with some initial AWS credentials.
//
// For example, if the given profile names a section with literal AWS
// credentials, the chain would look like:
// credentials → done
//
// A more complicated chain could look like:
// credentials → assume role → assume role → done
func Chain(cfg *config.Config, profile string) (*sts.Credentials, []Transformer, error) {
	return chain(cfg, profile, map[string]struct{}{
		profile: {},
	})
}

func chain(cfg *config.Config, profile string, seen map[string]struct{}) (*sts.Credentials, []Transformer, error) {
	// Look up the named profile. Maybe it's a user? Maybe it's a role?
	maybeUser, maybeRole, maybeSession, maybeFederate, err := cfg.Profile(profile)
	if err != nil {
		return nil, nil, chainError{
			profile: profile,
			err:     err,
		}
	}

	switch {
	case maybeUser != nil:
		// We have found a user with credentials, so no more recursive
		// searching is needed.
		creds := sts.Credentials{
			AccessKeyId:     aws.String(maybeUser.AWSAccessKeyID),
			SecretAccessKey: aws.String(maybeUser.AWSSecretAccessKey),
			SessionToken:    aws.String(maybeUser.AWSSessionToken),
		}

		return &creds, nil, nil

	case maybeFederate != nil:
		// We have found a federate. Check that we have not visited this profile
		// already, as that would mean that there is a circular profile
		// reference (mis)configured.
		if _, found := seen[maybeFederate.SourceProfile]; found {
			return nil, nil, chainError{
				profile: profile,
				err:     fmt.Errorf("recursive profile"),
			}
		}
		seen[maybeFederate.SourceProfile] = struct{}{}

		// Recursively follow the source profile reference, to walk the profile
		// "chain".
		creds, chain, err := chain(cfg, maybeFederate.SourceProfile, seen)
		if err != nil {
			return nil, nil, chainError{
				profile: profile,
				err:     err,
			}
		}

		// Create a get-federation-token transformer for this profile, and add
		// it to the chain.
		transform := FederationTokenTransform{
			Federate: maybeFederate,
		}
		chain = append(chain, transform)

		return creds, chain, err

	case maybeRole != nil:
		// We have found a role. Check that we have not visited this profile
		// already, as that would mean that there is a circular profile
		// reference (mis)configured.
		if _, found := seen[maybeRole.SourceProfile]; found {
			return nil, nil, chainError{
				profile: profile,
				err:     fmt.Errorf("recursive profile"),
			}
		}
		seen[maybeRole.SourceProfile] = struct{}{}

		// Recursively follow the source profile reference, to walk the profile
		// "chain".
		creds, chain, err := chain(cfg, maybeRole.SourceProfile, seen)
		if err != nil {
			return nil, nil, chainError{
				profile: profile,
				err:     err,
			}
		}

		// Create an assume-role transformer for this profile, and add it to
		// the chain.
		transform := AssumeRoleTransform{
			Role: maybeRole,
		}
		chain = append(chain, transform)

		return creds, chain, err

	case maybeSession != nil:
		// We have found a session. Check that we have not visited this profile
		// already, as that would mean that there is a circular profile
		// reference (mis)configured.
		if _, found := seen[maybeSession.SourceProfile]; found {
			return nil, nil, chainError{
				profile: profile,
				err:     fmt.Errorf("recursive profile"),
			}
		}
		seen[maybeSession.SourceProfile] = struct{}{}

		// Recursively follow the source profile reference, to walk the profile
		// "chain".
		creds, chain, err := chain(cfg, maybeSession.SourceProfile, seen)
		if err != nil {
			return nil, nil, chainError{
				profile: profile,
				err:     err,
			}
		}

		// Create an assume-role transformer for this profile, and add it to
		// the chain.
		transform := SessionTokenTransform{
			Session: maybeSession,
		}
		chain = append(chain, transform)

		return creds, chain, err
	}

	panic("internal error constructing profile chain")
}

type chainError struct {
	profile string
	err     error
}

func (e chainError) Error() string {
	return "profile chain " + e.error()
}

func (e chainError) error() string {
	switch cerr := e.err.(type) {
	case chainError:
		return e.profile + " → " + cerr.error()
	default:
		return e.profile + ": " + e.err.Error()
	}
}
