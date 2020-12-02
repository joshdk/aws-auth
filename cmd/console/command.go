// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package console

import (
	"fmt"

	"github.com/joshdk/aws-auth/config"
	"github.com/joshdk/aws-auth/console"
	"github.com/joshdk/aws-auth/transformers"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

// Command defines the aws-auth console command.
//
// $ aws-auth console
func Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "console",
		Short: "aws-auth console - Generate an AWS Console login URL",

		RunE: func(cmd *cobra.Command, args []string) error {
			flagBrowser, _ := cmd.Flags().GetBool("browser")
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

			// Generate an AWS Console login URL.
			url, err := console.GenerateLoginURL(endCreds)
			if err != nil {
				return err
			}

			// Print the login url.
			fmt.Println(url)

			// If the --browser flag is used, launch the url with the default
			// browser.
			if flagBrowser {
				return browser.OpenURL(url.String())
			}

			return nil
		},
	}

	cmd.Flags().BoolP("browser", "b", false, "open url with default browser")

	return cmd
}
