package agents

import (
	"embed"
)

// embed docs folder
//
//go:embed docs/*
var docsEmbedFS embed.FS
