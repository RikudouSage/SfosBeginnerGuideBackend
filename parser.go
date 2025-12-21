package main

import (
	"embed"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

type ContentItemLink struct {
	Link  string `json:"link"`
	Title string `json:"title"`
}

type ContentItemSection struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ContentItemMeta struct {
	Title   string   `yaml:"title" json:"title"`
	Links   []string `yaml:"links" json:"links"`
	Actions []string `yaml:"actions" json:"actions"`
}

type ContentItem struct {
	Meta     *ContentItemMeta      `json:"meta"`
	Content  string                `json:"content"`
	Sections []*ContentItemSection `json:"sections,omitempty"`
	Links    []*ContentItemLink    `json:"links,omitempty"`
}

type appMdParser struct {
	root     *embed.FS
	markdown goldmark.Markdown
	cache    *cache
}

func newParser(rootFs *embed.FS, markdown goldmark.Markdown) *appMdParser {
	return &appMdParser{
		root:     rootFs,
		markdown: markdown,
		cache:    newCache(5 * time.Minute),
	}
}

func (receiver *appMdParser) parseByPath(targetPath string) (*ContentItem, error) {
	targetPath = receiver.normalizePath(targetPath, "")

	if item, ok := receiver.cache.get(targetPath); ok {
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

	receiver.cache.set(targetPath, item)

	return item, nil
}

func (receiver *appMdParser) normalizePath(target string, relativeTo string) string {
	return normalizePath(target, relativeTo)
}

func (receiver *appMdParser) parse(content []byte, currentFile string) (*ContentItem, error) {
	result := &ContentItem{Meta: &ContentItemMeta{}}

	ctx := parser.NewContext()
	ctx.Set(linkResolverContextKey, currentFile)
	_ = receiver.markdown.Parser().Parse(text.NewReader(content), parser.WithContext(ctx))

	sectionsData, _ := ctx.Get(sectionContextKey).(*sectionInfo)
	if sectionsData == nil {
		return nil, fmt.Errorf("section splitter did not run")
	}

	introHTML, err := renderNodes(receiver.markdown, content, sectionsData.intro)
	if err != nil {
		return nil, fmt.Errorf("failed to render intro content: %w", err)
	}
	result.Content = introHTML

	for _, section := range sectionsData.sections {
		sectionHTML, err := renderNodes(receiver.markdown, content, section.nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to render section %s: %w", section.title, err)
		}

		result.Sections = append(result.Sections, &ContentItemSection{
			Title:   section.title,
			Content: sectionHTML,
		})
	}

	err = receiver.parseMetadata(ctx, result.Meta)
	if err != nil {
		return nil, fmt.Errorf("failed parsing metadata: %w", err)
	}
	err = receiver.parseLinks(result, currentFile)
	if err != nil {
		return nil, fmt.Errorf("failed parsing links: %w", err)
	}

	return result, nil
}

func (receiver *appMdParser) parseMetadata(ctx parser.Context, meta *ContentItemMeta) error {
	metadata := frontmatter.Get(ctx)
	return metadata.Decode(meta)
}

func (receiver *appMdParser) parseLinks(result *ContentItem, currentFile string) error {
	if len(result.Meta.Links) == 0 {
		return nil
	}

	for _, rawLink := range result.Meta.Links {
		targetFile := strings.TrimPrefix(receiver.normalizePath(rawLink, currentFile), "docs/")
		item, err := receiver.parseByPath(targetFile)
		if err != nil {
			return fmt.Errorf("failed parsing link %s: %w", rawLink, err)
		}

		result.Links = append(result.Links, &ContentItemLink{
			Link:  targetFile,
			Title: item.Meta.Title,
		})
	}

	return nil
}
