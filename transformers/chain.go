package transformers

import (
	"fmt"
	"strings"

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
	maybeUser, maybeRole := cfg.Profile(profile)

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

	case maybeRole != nil:
		// We have found a role. Check that we have not visited this role
		// already, as that would mean that there is a circular profile
		// reference (mis)configured.
		if _, found := seen[maybeRole.SourceProfile]; found {
			return nil, nil, fmt.Errorf("profile %s references recursive source profile %s", profile, maybeRole.SourceProfile)
		}
		seen[maybeRole.SourceProfile] = struct{}{}

		// Recursively follow the source profile reference, to walk the profile
		// "chain".
		creds, chain, err := chain(cfg, maybeRole.SourceProfile, seen)
		if err != nil {
			if strings.HasPrefix(err.Error(), "unknown source profile") {
				return nil, nil, fmt.Errorf("profile %s references %s", profile, err.Error())
			}
			return nil, nil, err
		}

		// Create an assume-role transformer for this profile, and add it to
		// the chain.
		transform := AssumeRoleTransform{
			Role: maybeRole,
		}
		chain = append(chain, transform)

		return creds, chain, err

	default:
		// We have neither a user nor a role, so something went wrong. Either
		// an unknown profile was requested, or a profile with misconfigured
		// content was encountered.
		return nil, nil, fmt.Errorf("unknown source profile %s", profile)
	}
}
