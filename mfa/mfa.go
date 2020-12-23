// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package mfa

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	ykman "github.com/joshdk/ykmango"
)

// Prompt requests that the user enter an MFA code. If a Yubikey slot name is
// given, a code is directly requested from the device, and may require
// touching the Yubikey.
func Prompt(serial, message, yubikeySlot string) (string, error) {
	// Print a prompt message so that the user knows what to do.
	if message != "" {
		fmt.Fprintf(os.Stderr, "%s ", message)
	} else {
		fmt.Fprintf(os.Stderr, "Enter MFA code for %s: ", serial)
	}

	if yubikeySlot != "" {
		// Since the user will not hit enter, print an extra newline.
		defer fmt.Fprintln(os.Stderr, "")

		// Generate an MFA code from the Yubikey.
		return ykman.Generate(yubikeySlot)
	}

	// Read code from the line that the user types in.
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(code), err
}
