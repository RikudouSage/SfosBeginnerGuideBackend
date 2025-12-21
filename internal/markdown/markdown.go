package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"go.abhg.dev/goldmark/frontmatter"
)

func New() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Strikethrough,
			&frontmatter.Extender{},
			newLinkResolver(),
			newSectionSplitter(),
		),
	)
}
