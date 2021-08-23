package webserver

import "embed"

//go:embed embedded/*.html
var templateFS embed.FS

//go:embed embedded/*.js
//go:embed embedded/*.css
var staticFS embed.FS
