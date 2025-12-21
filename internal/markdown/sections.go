package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var SectionContextKey = parser.NewContextKey()

type SectionInfo struct {
	Intro    []ast.Node
	Sections []SectionRecord
}

type SectionRecord struct {
	Title string
	Nodes []ast.Node
}

func newSectionSplitter() goldmark.Extender { return sectionSplitter{} }

type sectionSplitter struct{}

func (sectionSplitter) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(sectionTransformer{}, 200),
		),
	)
}

type sectionTransformer struct{}

func (sectionTransformer) Transform(node *ast.Document, reader text.Reader, parserContext parser.Context) {
	info := &SectionInfo{}

	var currentTitle string
	var currentNodes []ast.Node

	for currentNode := node.FirstChild(); currentNode != nil; currentNode = currentNode.NextSibling() {
		if heading, ok := currentNode.(*ast.Heading); ok && heading.Level == 2 {
			if currentTitle == "" {
				info.Intro = append(info.Intro, currentNodes...)
			} else {
				info.Sections = append(info.Sections, SectionRecord{
					Title: currentTitle,
					Nodes: currentNodes,
				})
			}

			currentTitle = headingText(heading, reader.Source())
			currentNodes = nil
			continue
		}

		currentNodes = append(currentNodes, currentNode)
	}

	if currentTitle == "" {
		info.Intro = append(info.Intro, currentNodes...)
	} else {
		info.Sections = append(info.Sections, SectionRecord{
			Title: currentTitle,
			Nodes: currentNodes,
		})
	}

	parserContext.Set(SectionContextKey, info)
}

func headingText(heading *ast.Heading, source []byte) string {
	var buf bytes.Buffer
	for n := heading.FirstChild(); n != nil; n = n.NextSibling() {
		buf.Write(n.Text(source))
	}
	return buf.String()
}

func RenderNodes(md goldmark.Markdown, source []byte, nodes []ast.Node) (string, error) {
	var buf bytes.Buffer
	for _, n := range nodes {
		if err := md.Renderer().Render(&buf, source, n); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}
