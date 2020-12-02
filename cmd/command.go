// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package cmd

import (
	"fmt"
	"os"

	"github.com/joshdk/aws-auth/cmd/console"
	"github.com/joshdk/aws-auth/config"
	"github.com/joshdk/aws-auth/transformers"
	"github.com/spf13/cobra"
)

// Command defines the aws-auth CLI root.
//
// $ aws-auth
func Command(version, date string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "aws-auth",
		Short:         "aws-auth - Manage AWS credential for a range of workflows",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			flagProfile, _ := cmd.Flags().GetString("profile")

			// Load and parse the AWS config files.
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Find a chain of transforms for obtaining profile credentials.
			startCreds, transforms, err := transformers.Chain(cfg, flagProfile)
			if err != nil {
				return err
			}

			// Make all of the transforms needed to obtain those credentials.
			endCreds, err := transformers.Transform(startCreds, transforms)
			if err != nil {
				return err
			}

			// Enrich credentials with identity information.
			identity, err := transformers.Enrich(endCreds)
			if err != nil {
				return err
			}

			// Print environment variables for our new identity.
			for key, value := range identity.Env() {
				fmt.Printf("export %s=%q\n", key, value)
			}

			return nil
		},
	}

	cmd.SetVersionTemplate(versionTemplate(version, date))

	cmd.PersistentFlags().StringP("profile", "p", "default", "AWS profile to target")

	cmd.AddCommand(
		console.Command(),
	)

	return cmd
}

// Execute handles the CLI and runs it to completion. This function does not
// return.
func Execute(version, date string) {
	if err := Command(version, date).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "aws-auth: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// versionTemplate formats the command output when using the --version flag.
func versionTemplate(version, date string) string {
	return fmt.Sprintf(
		"version: %s\n"+
			"date:    %s\n"+
			"author:  Josh Komoroske\n"+
			"license: MIT\n"+
			"github:  https://github.com/joshdk/aws-auth\n",
		version, date,
	)
}
