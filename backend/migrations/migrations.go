// Package migrations embeds the SQL migration files so they ship inside the
// binary. This lets `migrate` run identically on a dev laptop and in a
// container without mounting a migrations directory.
package migrations

import "embed"

// FS holds all *.sql migration files.
//
//go:embed *.sql
var FS embed.FS
