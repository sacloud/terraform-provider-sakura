// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package ver

import "fmt"

var (
	// Version app version
	Version = "3.0.0-rc8"
	// Revision git commit short commithash
	Revision = "xxxxxx" // set on build
)

// FullVersion return sackerel full version text
func FullVersion() string {
	return fmt.Sprintf("%s, build %s", Version, Revision)
}
