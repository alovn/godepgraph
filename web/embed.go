package web

import (
	"embed"
	_ "embed"
)

//go:embed dist/*
var distFiles embed.FS
