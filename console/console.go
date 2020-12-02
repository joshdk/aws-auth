package console

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
)

// GenerateLoginURL takes the given sts.Credentials and generates a url.URL
// that can be used to login to the AWS Console.
// https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_enable-console-custom-url.html
func GenerateLoginURL(creds *sts.Credentials) (*url.URL, error) {
	federationURL, _ := url.Parse("https://signin.aws.amazon.com/federation")

	type requestCredentials struct {
		SessionID    string `json:"sessionId"`
		SessionKey   string `json:"sessionKey"`
		SessionToken string `json:"sessionToken"`
	}

	rq := requestCredentials{
		SessionID:    aws.StringValue(creds.AccessKeyId),
		SessionKey:   aws.StringValue(creds.SecretAccessKey),
		SessionToken: aws.StringValue(creds.SessionToken),
	}

	data, err := json.Marshal(rq)
	if err != nil {
		return nil, err
	}

	// Set proper URL query parameters for request.
	values := url.Values{
		"Action":          []string{"getSigninToken"},
		"SessionDuration": []string{"3600"},
		"Session":         []string{string(data)},
	}
	federationURL.RawQuery = values.Encode()

	// Perform the actual API request.
	resp, err := http.Get(federationURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Verify that we received a 200 OK.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}

	// Extract a sign-in token from the response JSON.
	token, err := extractToken(resp.Body)
	if err != nil {
		return nil, err
	}

	// Format sign-in URL and return it!
	return signinURL(token), nil
}

// extractToken parses the response JSON from a getSigninToken request and
// returns the contained sign-in token.
func extractToken(reader io.Reader) (string, error) {
	type response struct {
		SigninToken string `json:"SigninToken"`
	}

	var resp response
	if err := json.NewDecoder(reader).Decode(&resp); err != nil {
		return "", err
	}

	return resp.SigninToken, nil
}

// signinURL formats a AWS console login url using the given sign-in token.
func signinURL(token string) *url.URL {
	result, _ := url.Parse("https://signin.aws.amazon.com/federation")

	values := url.Values{
		"Action":      []string{"login"},
		"Destination": []string{"https://console.aws.amazon.com/"},
		"SigninToken": []string{token},
	}
	result.RawQuery = values.Encode()

	return result
}
