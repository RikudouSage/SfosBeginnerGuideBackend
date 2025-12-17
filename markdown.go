package main

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"go.abhg.dev/goldmark/frontmatter"
)

var markdown = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Strikethrough,
		&frontmatter.Extender{},
	),
)
