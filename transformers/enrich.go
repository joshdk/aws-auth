package transformers

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

type Identity struct {
	ARN             string
	AccessKeyID     string
	AccountID       string
	Expiration      time.Time
	SecretAccessKey string
	SessionToken    string
}

// Env returns the Identity contents as a map of key/value pairs, intended to
// be used as environment variables. Blank value are omitted.
func (i Identity) Env() map[string]string {
	vars := map[string]string{
		"AWS_ACCESS_KEY_ID":     i.AccessKeyID,
		"AWS_ACCOUNT_ID":        i.AccountID,
		"AWS_ARN":               i.ARN,
		"AWS_SECRET_ACCESS_KEY": i.SecretAccessKey,
	}

	// IAM keys do not have an expiration.
	if i.Expiration != (time.Time{}) {
		vars["AWS_EXPIRATION"] = i.Expiration.String()
	}

	// IAM keys do not have an associated session token.
	if i.SessionToken != "" {
		vars["AWS_SESSION_TOKEN"] = i.SessionToken
	}

	return vars
}

// Enrich combines the given sts.Credentials with information about their
// associated principal for convenience.
func Enrich(creds *sts.Credentials) (*Identity, error) {
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
	output, err := sts.New(sess).GetCallerIdentity(nil)
	if err != nil {
		return nil, err
	}

	return &Identity{
		ARN:             aws.StringValue(output.Arn),
		AccessKeyID:     aws.StringValue(creds.AccessKeyId),
		AccountID:       aws.StringValue(output.Account),
		Expiration:      aws.TimeValue(creds.Expiration),
		SecretAccessKey: aws.StringValue(creds.SecretAccessKey),
		SessionToken:    aws.StringValue(creds.SessionToken),
	}, nil
}
