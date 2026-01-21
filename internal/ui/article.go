package ui

import (
	"fmt"
	"strings"

	"github.com/tmustier/economist-cli/internal/article"
)

func RenderArticleHeader(art *article.Article, styles ArticleStyles) string {
	var sb strings.Builder
	if art.Overtitle != "" {
		sb.WriteString(styles.Overtitle.Render(art.Overtitle))
		sb.WriteString("\n")
	}
	if art.Title != "" {
		sb.WriteString(styles.Title.Render(art.Title))
		sb.WriteString("\n")
	}
	if art.Subtitle != "" {
		sb.WriteString(styles.Subtitle.Render(art.Subtitle))
		sb.WriteString("\n")
	}
	if art.DateLine != "" {
		sb.WriteString(styles.Date.Render(art.DateLine))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	sb.WriteString(styles.Rule.Render("--------"))
	sb.WriteString("\n\n")
	return sb.String()
}

func ArticleBodyMarkdown(art *article.Article) string {
	var sb strings.Builder
	if art.Content != "" {
		sb.WriteString(art.Content)
	}
	return sb.String()
}

func ArticleFooter(art *article.Article, styles ArticleStyles) string {
	var sb strings.Builder
	sb.WriteString("\n\n")
	sb.WriteString(styles.Rule.Render("--------"))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("🔗 %s\n", art.URL))
	return sb.String()
}

func HighlightTrailingMarker(text string, styles ArticleStyles) string {
	idx := strings.LastIndex(text, "■")
	if idx == -1 {
		return text
	}

	marker := styles.Title.Render("■")
	return text[:idx] + marker + text[idx+len("■"):]
}
