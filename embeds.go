package admin

import "embed"

//go:embed templates
var templateEmbededFS embed.FS

//go:embed static
var staticEmbededFS embed.FS
