package content

import (
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"

	"SfosBeginnerGuide/internal/cache"
	"SfosBeginnerGuide/internal/markdown"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

type Parser interface {
	ParseByPath(path string) (*Item, error)
}

type MarkdownParser struct {
	root     fs.FS
	markdown goldmark.Markdown
	cache    cache.Store[*Item]
}

func NewMarkdownParser(root fs.FS, md goldmark.Markdown, cacheStore cache.Store[*Item]) *MarkdownParser {
	return &MarkdownParser{
		root:     root,
		markdown: md,
		cache:    cacheStore,
	}
}

func NewCachedMarkdownParser(root fs.FS, md goldmark.Markdown, ttl time.Duration) *MarkdownParser {
	return NewMarkdownParser(root, md, cache.NewTTL[*Item](ttl))
}

func (receiver *MarkdownParser) ParseByPath(targetPath string) (*Item, error) {
	targetPath = markdown.NormalizePath(targetPath, "")

	if item, ok := receiver.cache.Get(targetPath); ok {
		return item, nil
	}

	file, err := receiver.root.Open(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", targetPath, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", targetPath, err)
	}

	item, err := receiver.parse(content, targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", targetPath, err)
	}

	receiver.cache.Set(targetPath, item)

	return item, nil
}

func (receiver *MarkdownParser) parse(content []byte, currentFile string) (*Item, error) {
	result := &Item{Meta: &Meta{}}

	ctx := parser.NewContext()
	ctx.Set(markdown.LinkResolverContextKey, currentFile)
	_ = receiver.markdown.Parser().Parse(text.NewReader(content), parser.WithContext(ctx))

	sectionsData, _ := ctx.Get(markdown.SectionContextKey).(*markdown.SectionInfo)
	if sectionsData == nil {
		return nil, fmt.Errorf("section splitter did not run")
	}

	introHTML, err := markdown.RenderNodes(receiver.markdown, content, sectionsData.Intro)
	if err != nil {
		return nil, fmt.Errorf("failed to render intro content: %w", err)
	}
	result.Content = introHTML

	for _, section := range sectionsData.Sections {
		sectionHTML, err := markdown.RenderNodes(receiver.markdown, content, section.Nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to render section %s: %w", section.Title, err)
		}

		result.Sections = append(result.Sections, &Section{
			Title:   section.Title,
			Content: sectionHTML,
		})
	}

	if err := receiver.parseMetadata(ctx, result.Meta); err != nil {
		return nil, fmt.Errorf("failed parsing metadata: %w", err)
	}
	if err := receiver.parseLinks(result, currentFile); err != nil {
		return nil, fmt.Errorf("failed parsing links: %w", err)
	}

	return result, nil
}

func (receiver *MarkdownParser) parseMetadata(ctx parser.Context, meta *Meta) error {
	metadata := frontmatter.Get(ctx)
	return metadata.Decode(meta)
}

func (receiver *MarkdownParser) parseLinks(result *Item, currentFile string) error {
	if len(result.Meta.Links) == 0 {
		return nil
	}

	for _, rawLink := range result.Meta.Links {
		targetFile := strings.TrimPrefix(markdown.NormalizePath(rawLink, currentFile), "docs/")
		item, err := receiver.ParseByPath(targetFile)
		if err != nil {
			return fmt.Errorf("failed parsing link %s: %w", rawLink, err)
		}

		result.Links = append(result.Links, &Link{
			Link:  targetFile,
			Title: item.Meta.Title,
		})
	}

	return nil
}
