// Package migrations embeds the golang-migrate SQL files so the API binary can
// apply the schema itself (on boot or via `webond migrate`) without shipping a
// separate migration tool or files next to the distroless image.
package migrations

import "embed"

// FS holds every *.up.sql / *.down.sql migration in this directory.
//
//go:embed *.sql
var FS embed.FS
