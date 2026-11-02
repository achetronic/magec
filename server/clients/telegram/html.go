// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package telegram

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/achetronic/magec/server/clients/msgutil"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// markdownParser parses the CommonMark + GFM dialect LLMs usually write.
var markdownParser = goldmark.New(goldmark.WithExtensions(extension.GFM))

// codeLanguagePattern restricts fenced code languages to safe class names.
var codeLanguagePattern = regexp.MustCompile(`^[A-Za-z0-9_+#.-]+$`)

// htmlTagPattern matches the tags emitted by the Telegram HTML renderer.
var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// markdownToHTML converts agent Markdown into Telegram HTML
// (parse_mode "HTML") split into messages of at most maxLen characters.
//
// Telegram's legacy Markdown parse mode rejects common LLM output such as
// snake_case identifiers, unpaired asterisks, "**bold**" or "# headings".
// Parsing the text with a real Markdown parser and escaping everything that is
// not formatting avoids those errors while keeping code blocks, emphasis and
// links. Splits happen between Markdown blocks, so every chunk is valid HTML
// on its own; oversized code blocks are split by lines.
func markdownToHTML(md string, maxLen int) []string {
	if maxLen <= 0 {
		maxLen = msgutil.TelegramMaxMessageLength
	}

	src := []byte(md)
	doc := markdownParser.Parser().Parse(text.NewReader(src))
	r := &htmlRenderer{src: src}

	var fragments []string
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		fragments = append(fragments, r.fragments(n, maxLen)...)
	}

	var chunks []string
	current := ""
	for _, frag := range fragments {
		if frag == "" {
			continue
		}
		if current == "" {
			current = frag
			continue
		}
		if textLength(current+"\n\n"+frag) <= maxLen {
			current += "\n\n" + frag
			continue
		}
		chunks = append(chunks, current)
		current = frag
	}
	if current != "" {
		chunks = append(chunks, current)
	}
	return chunks
}

// htmlToPlain strips Telegram HTML tags and entities, returning the
// text the user would see.
func htmlToPlain(s string) string {
	return html.UnescapeString(htmlTagPattern.ReplaceAllString(s, ""))
}

// textLength returns the length Telegram counts against its message
// limit: visible text (after entity parsing) in UTF-16 code units.
func textLength(s string) int {
	return len(utf16.Encode([]rune(htmlToPlain(s))))
}

// htmlRenderer renders a goldmark AST into Telegram-supported HTML tags.
type htmlRenderer struct {
	src []byte
}

// fragments renders a top-level block. When the result does not fit in a
// single message it is broken into smaller, independently valid fragments.
func (r *htmlRenderer) fragments(n ast.Node, maxLen int) []string {
	out := r.block(n, false)
	if textLength(out) <= maxLen {
		return []string{out}
	}

	switch n := n.(type) {
	case *ast.FencedCodeBlock:
		return splitCodeBlock(r.codeLanguage(n), r.lines(n), maxLen)
	case *ast.CodeBlock:
		return splitCodeBlock("", r.lines(n), maxLen)
	case *ast.List:
		var items []string
		for i, item := 0, n.FirstChild(); item != nil; i, item = i+1, item.NextSibling() {
			rendered := r.listItem(n, item, i, "", false)
			if textLength(rendered) <= maxLen {
				items = append(items, rendered)
				continue
			}
			items = append(items, splitPlain(htmlToPlain(rendered), maxLen)...)
		}
		return items
	}
	return splitPlain(htmlToPlain(out), maxLen)
}

// block renders a block node. inQuote avoids nested blockquotes, which
// Telegram does not support.
func (r *htmlRenderer) block(n ast.Node, inQuote bool) string {
	switch n := n.(type) {
	case *ast.Paragraph, *ast.TextBlock:
		return r.inlines(n)
	case *ast.Heading:
		return "<b>" + r.inlines(n) + "</b>"
	case *ast.ThematicBreak:
		return "──────────"
	case *ast.FencedCodeBlock:
		return codeBlock(r.codeLanguage(n), strings.Join(r.lines(n), ""))
	case *ast.CodeBlock:
		return codeBlock("", strings.Join(r.lines(n), ""))
	case *ast.HTMLBlock:
		return msgutil.EscapeHTML(strings.TrimRight(strings.Join(r.lines(n), ""), "\n"))
	case *ast.Blockquote:
		inner := r.children(n, "\n", true)
		if inQuote {
			return inner
		}
		return "<blockquote>" + inner + "</blockquote>"
	case *ast.List:
		return r.list(n, "", inQuote)
	case *east.Table:
		return r.table(n)
	}
	return r.children(n, "\n\n", inQuote)
}

// children renders the block children of n joined by sep.
func (r *htmlRenderer) children(n ast.Node, sep string, inQuote bool) string {
	var parts []string
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if s := r.block(c, inQuote); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, sep)
}

// list renders a list as bullet ("•") or numbered lines. Nested lists are
// indented under their parent item.
func (r *htmlRenderer) list(l *ast.List, indent string, inQuote bool) string {
	var lines []string
	for i, item := 0, l.FirstChild(); item != nil; i, item = i+1, item.NextSibling() {
		lines = append(lines, r.listItem(l, item, i, indent, inQuote))
	}
	return strings.Join(lines, "\n")
}

// listItem renders the i-th item of l with its marker.
func (r *htmlRenderer) listItem(l *ast.List, item ast.Node, i int, indent string, inQuote bool) string {
	marker := "• "
	if l.IsOrdered() {
		marker = fmt.Sprintf("%d. ", l.Start+i)
	}

	var parts []string
	for c := item.FirstChild(); c != nil; c = c.NextSibling() {
		if nested, ok := c.(*ast.List); ok {
			parts = append(parts, r.list(nested, indent+"    ", inQuote))
			continue
		}
		if s := r.block(c, inQuote); s != "" {
			parts = append(parts, s)
		}
	}
	return indent + marker + strings.Join(parts, "\n")
}

// table renders a GFM table as aligned plain text inside <pre>, since
// Telegram has no table support.
func (r *htmlRenderer) table(t *east.Table) string {
	var rows [][]string
	for row := t.FirstChild(); row != nil; row = row.NextSibling() {
		var cells []string
		for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
			cells = append(cells, htmlToPlain(r.inlines(cell)))
		}
		rows = append(rows, cells)
	}

	var widths []int
	for _, row := range rows {
		for i, cell := range row {
			if i >= len(widths) {
				widths = append(widths, 0)
			}
			widths[i] = max(widths[i], len([]rune(cell)))
		}
	}

	var lines []string
	for i, row := range rows {
		padded := make([]string, len(row))
		for j, cell := range row {
			padded[j] = cell + strings.Repeat(" ", widths[j]-len([]rune(cell)))
		}
		lines = append(lines, strings.TrimRight(strings.Join(padded, " | "), " "))
		if i == 0 {
			seps := make([]string, len(widths))
			for j, w := range widths {
				seps[j] = strings.Repeat("-", w)
			}
			lines = append(lines, strings.Join(seps, "-+-"))
		}
	}
	return "<pre>" + msgutil.EscapeHTML(strings.Join(lines, "\n")) + "</pre>"
}

// inlines renders the inline children of n.
func (r *htmlRenderer) inlines(n ast.Node) string {
	var sb strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		sb.WriteString(r.inline(c))
	}
	return sb.String()
}

func (r *htmlRenderer) inline(n ast.Node) string {
	switch n := n.(type) {
	case *ast.Text:
		value := n.Segment.Value(r.src)
		if !n.IsRaw() {
			value = util.ResolveEntityNames(util.ResolveNumericReferences(util.UnescapePunctuations(value)))
		}
		s := msgutil.EscapeHTML(string(value))
		if n.SoftLineBreak() || n.HardLineBreak() {
			s += "\n"
		}
		return s
	case *ast.String:
		return msgutil.EscapeHTML(string(n.Value))
	case *ast.CodeSpan:
		var sb strings.Builder
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if t, ok := c.(*ast.Text); ok {
				sb.Write(t.Segment.Value(r.src))
			}
		}
		code := strings.ReplaceAll(sb.String(), "\n", " ")
		return "<code>" + msgutil.EscapeHTML(code) + "</code>"
	case *ast.Emphasis:
		tag := "i"
		if n.Level >= 2 {
			tag = "b"
		}
		return "<" + tag + ">" + r.inlines(n) + "</" + tag + ">"
	case *east.Strikethrough:
		return "<s>" + r.inlines(n) + "</s>"
	case *ast.Link:
		return link(string(n.Destination), r.inlines(n))
	case *ast.AutoLink:
		return link(string(n.URL(r.src)), msgutil.EscapeHTML(string(n.Label(r.src))))
	case *ast.Image:
		return link(string(n.Destination), r.inlines(n))
	case *ast.RawHTML:
		var sb strings.Builder
		for i := 0; i < n.Segments.Len(); i++ {
			seg := n.Segments.At(i)
			sb.Write(seg.Value(r.src))
		}
		return msgutil.EscapeHTML(sb.String())
	case *east.TaskCheckBox:
		if n.IsChecked {
			return "☑ "
		}
		return "☐ "
	}
	return r.inlines(n)
}

// lines returns the raw source lines of a leaf block.
func (r *htmlRenderer) lines(n ast.Node) []string {
	lines := n.Lines()
	out := make([]string, 0, lines.Len())
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		out = append(out, string(seg.Value(r.src)))
	}
	return out
}

func (r *htmlRenderer) codeLanguage(n *ast.FencedCodeBlock) string {
	lang := string(n.Language(r.src))
	if !codeLanguagePattern.MatchString(lang) {
		return ""
	}
	return lang
}

// codeBlock renders a <pre> block, tagging the language when known.
func codeBlock(lang, code string) string {
	code = msgutil.EscapeHTML(strings.TrimRight(code, "\n"))
	if lang == "" {
		return "<pre>" + code + "</pre>"
	}
	return `<pre><code class="language-` + lang + `">` + code + "</code></pre>"
}

// splitCodeBlock splits a code block by lines into several blocks that each
// fit in maxLen. A single line longer than maxLen is hard-cut.
func splitCodeBlock(lang string, lines []string, maxLen int) []string {
	var blocks []string
	current := ""
	for _, line := range lines {
		if current != "" && textLength(codeBlock(lang, current+line)) > maxLen {
			blocks = append(blocks, codeBlock(lang, current))
			current = ""
		}
		if current == "" && textLength(codeBlock(lang, line)) > maxLen {
			for _, piece := range splitPlain(line, maxLen) {
				blocks = append(blocks, codeBlock(lang, htmlToPlain(piece)))
			}
			continue
		}
		current += line
	}
	if current != "" {
		blocks = append(blocks, codeBlock(lang, current))
	}
	return blocks
}

// splitPlain splits plain text into escaped HTML pieces of at most maxLen UTF-16
// code units, losing formatting. Only used for blocks too large to send whole.
func splitPlain(s string, maxLen int) []string {
	limit := maxLen
	for {
		pieces := msgutil.SplitMessage(s, limit)
		if piecesFitUTF16(pieces, maxLen) || limit <= 1 {
			out := make([]string, len(pieces))
			for i, p := range pieces {
				out[i] = msgutil.EscapeHTML(p)
			}
			return out
		}
		// SplitMessage limits by runes, but Telegram counts UTF-16 code units.
		// Characters outside the BMP (e.g. emoji) are 1 rune but 2 UTF-16 units,
		// so a rune-based limit can overestimate the real length. We halve the
		// rune limit until the UTF-16 length fits, or until we reach 1.
		limit /= 2
	}
}

// piecesFitUTF16 reports whether every piece, after HTML escaping, fits within
// maxLen UTF-16 code units as Telegram measures it.
func piecesFitUTF16(pieces []string, maxLen int) bool {
	for _, p := range pieces {
		if textLength(msgutil.EscapeHTML(p)) > maxLen {
			return false
		}
	}
	return true
}

// link renders an anchor for URL schemes Telegram accepts, or just the label.
func link(dest, label string) string {
	if label == "" {
		label = msgutil.EscapeHTML(dest)
	}
	lower := strings.ToLower(dest)
	for _, scheme := range []string{"http://", "https://", "tg://", "mailto:"} {
		if strings.HasPrefix(lower, scheme) {
			return `<a href="` + msgutil.EscapeHTML(dest) + `">` + label + `</a>`
		}
	}
	return label
}
