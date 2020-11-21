// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Command defines the aws-auth CLI root.
//
// $ aws-auth
func Command(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "aws-auth",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}

// Execute handles the CLI and runs it to completion. This function does not
// return.
func Execute(version string) {
	if err := Command(version).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "aws-auth: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
