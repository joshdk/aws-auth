// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package transformers

import (
	"github.com/aws/aws-sdk-go/service/sts"
)

// Transformer represents types that are able to trade a given sts.Credentials
// value for a new (derived) sts.Credentials value.
type Transformer interface {
	Transform(*sts.Credentials) (*sts.Credentials, error)
}

// Transform is a reduce-style operation. The given sts.Credentials are passed
// to the first Transformer, the result of which is passed to the second, and
// so on.
func Transform(credentials *sts.Credentials, transformers []Transformer) (*sts.Credentials, error) {
	for _, transformer := range transformers {
		newCredentials, err := transformer.Transform(credentials)
		if err != nil {
			return nil, err
		}
		credentials = newCredentials
	}
	return credentials, nil
}
