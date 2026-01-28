// migrations.go (at the project root)
package jimu

import "embed"

//go:embed migrations/*.sql
var MigrationFiles embed.FS
