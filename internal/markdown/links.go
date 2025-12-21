package markdown

import (
	"net/url"
	"path"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var LinkResolverContextKey = parser.NewContextKey()

func newLinkResolver() goldmark.Extender { return linkResolver{} }

type linkResolver struct{}

func (linkResolver) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(linkResolverTransformer{}, 150),
		),
	)
}

type linkResolverTransformer struct{}

func (linkResolverTransformer) Transform(node *ast.Document, reader text.Reader, parserContext parser.Context) {
	currentFile, _ := parserContext.Get(LinkResolverContextKey).(string)
	if currentFile == "" {
		return
	}

	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		link, ok := n.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}

		destination := string(link.Destination)
		if ShouldResolveRelativeLink(destination) {
			resolved := ResolveAbsoluteMarkdownLink(destination, currentFile)
			if resolved != "" {
				link.Destination = []byte(resolved)
			}
		}

		return ast.WalkContinue, nil
	})
}

func NormalizePath(target string, relativeTo string) string {
	if strings.HasPrefix(target, "docs/") {
		target = target[len("docs/"):]
	}
	if strings.HasPrefix(relativeTo, "docs") {
		relativeTo = relativeTo[len("docs/"):]
	}

	target = strings.TrimPrefix(target, "/")

	base := relativeTo
	if strings.HasSuffix(base, ".md") {
		base = path.Dir(base)
	}

	target = path.Join(base, target)

	if !strings.HasSuffix(target, ".md") {
		if !strings.HasSuffix(target, "/") {
			target += "/"
		}
		target += "index.md"
	}

	target = path.Clean(target)

	return "docs/" + strings.TrimPrefix(target, "/")
}

func ShouldResolveRelativeLink(destination string) bool {
	if destination == "" {
		return false
	}
	if strings.HasPrefix(destination, "#") || strings.HasPrefix(destination, "/") || strings.HasPrefix(destination, "//") {
		return false
	}
	parsed, err := url.Parse(destination)
	if err == nil && parsed.Scheme != "" {
		return false
	}

	pathPart, _, _ := strings.Cut(destination, "#")
	if pathPart == "" {
		return false
	}

	if strings.HasSuffix(pathPart, "/") {
		return true
	}

	ext := path.Ext(pathPart)
	return ext == "" || ext == ".md"
}

func ResolveAbsoluteMarkdownLink(destination string, currentFile string) string {
	pathPart, fragment, _ := strings.Cut(destination, "#")
	if pathPart == "" {
		return ""
	}

	resolved := NormalizePath(pathPart, currentFile)
	resolved = "document:///" + strings.TrimPrefix(resolved, "docs/")

	if fragment != "" {
		resolved += "#" + fragment
	}

	return resolved
}
