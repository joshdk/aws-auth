// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package main

import "github.com/joshdk/aws-auth/cmd"

// version holds the version string. Replaced at build-time with -ldflags.
var version = "development"

// date holds the build date. Replaced at build-time with -ldflags.
var date = "-"

func main() {
	cmd.Execute(version, date)
}
