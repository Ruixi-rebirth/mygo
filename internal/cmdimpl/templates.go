package cmdimpl

import "embed"

//go:embed templates/binary/* templates/library/*
var templateFS embed.FS

type templateData struct {
	Name   string
	Module string
}
