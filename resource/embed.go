package resource

import "embed"

//go:embed template/* logo/*
var FS embed.FS
