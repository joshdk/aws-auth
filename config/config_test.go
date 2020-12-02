// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package config

import (
	"fmt"
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	const envVarHome = "testdata/empty"
	tests := []struct {
		env map[string]string
		err bool
	}{
		{
			err: true,
		},
		{
			env: map[string]string{
				EnvVarAWSConfigFile: "testdata/non-empty/.aws/config",
			},
		},
		{
			env: map[string]string{
				EnvVarAWSSharedCredentialsFile: "testdata/non-empty/.aws/credentials",
			},
		},
		{
			env: map[string]string{
				EnvVarAWSConfigFile:            "testdata/non-empty/.aws/config",
				EnvVarAWSSharedCredentialsFile: "testdata/non-empty/.aws/credentials",
			},
		},
		{
			env: map[string]string{
				"HOME": "testdata/non-empty",
			},
		},
		{
			env: map[string]string{
				"HOME":                         "testdata/non-empty",
				EnvVarAWSConfigFile:            "testdata/empty/.aws/config",
				EnvVarAWSSharedCredentialsFile: "testdata/empty/.aws/credentials",
			},
			err: true,
		},
	}

	for index, test := range tests {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			// Reset environment and recreate it for every test.
			os.Clearenv()
			os.Setenv("HOME", envVarHome)
			for key, value := range test.env {
				os.Setenv(key, value)
			}

			_, err := Load()
			switch {
			case err != nil && test.err:
				return
			case err != nil && !test.err:
				t.Fatalf("expected no error but got error %q", err)
			case err == nil && test.err:
				t.Fatalf("expected an error but got no error")
			}
		})
	}
}
