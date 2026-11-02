// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package telegram

import (
	"strings"
	"testing"

	"github.com/achetronic/magec/server/clients/msgutil"
)

func renderOne(t *testing.T, md string) string {
	t.Helper()
	chunks := markdownToHTML(md, msgutil.TelegramMaxMessageLength)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d: %q", len(chunks), chunks)
	}
	return chunks[0]
}

func TestMarkdownToHTML_Inline(t *testing.T) {
	cases := []struct {
		name string
		md   string
		want string
	}{
		{"bold", "**Ojo** aquí", "<b>Ojo</b> aquí"},
		{"italic", "*cursiva* y _otra_", "<i>cursiva</i> y <i>otra</i>"},
		{"strikethrough", "~~viejo~~", "<s>viejo</s>"},
		{"code span", "usa `terraform plan`", "usa <code>terraform plan</code>"},
		{"code span escapes", "`a < b && c > d`", "<code>a &lt; b &amp;&amp; c &gt; d</code>"},
		{"link", "[docs](https://example.com/a?b=1&c=2)", `<a href="https://example.com/a?b=1&amp;c=2">docs</a>`},
		{"autolink", "<https://example.com>", `<a href="https://example.com">https://example.com</a>`},
		{"relative link keeps label", "[fichero](./main.tf)", "fichero"},
		{"heading", "# Plan de cambios", "<b>Plan de cambios</b>"},
		{"raw html escaped", "usa <br> aquí", "usa &lt;br&gt; aquí"},
		{"backslash escape", `1\*2`, "1*2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := renderOne(t, tc.md); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// These inputs broke Telegram's legacy Markdown parse mode.
func TestMarkdownToHTML_LegacyMarkdownBreakers(t *testing.T) {
	cases := []struct {
		name string
		md   string
		want string
	}{
		{"snake_case", "el bucket my_state_bucket_prod", "el bucket my_state_bucket_prod"},
		{"glob", "borra los *.tfstate del bucket", "borra los *.tfstate del bucket"},
		{"unpaired asterisk", "precio 5 * 3", "precio 5 * 3"},
		{"lone backtick", "un ` suelto", "un ` suelto"},
		{"brackets", "array[0] y map[string]int", "array[0] y map[string]int"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := renderOne(t, tc.md); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMarkdownToHTML_CodeBlock(t *testing.T) {
	md := "Aplica esto:\n\n```hcl\nresource \"aws_s3_bucket\" \"state\" {\n  bucket = \"my_state_bucket\"\n}\n```\n\nY luego `terraform apply`."
	want := "Aplica esto:\n\n" +
		"<pre><code class=\"language-hcl\">resource \"aws_s3_bucket\" \"state\" {\n  bucket = \"my_state_bucket\"\n}</code></pre>\n\n" +
		"Y luego <code>terraform apply</code>."
	if got := renderOne(t, md); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestMarkdownToHTML_CodeBlockKeepsMarkup(t *testing.T) {
	md := "```bash\nkubectl get pods -o jsonpath='{.items[*].metadata.name}' | grep <pod> && echo **ok**\n```"
	want := "<pre><code class=\"language-bash\">kubectl get pods -o jsonpath='{.items[*].metadata.name}' | grep &lt;pod&gt; &amp;&amp; echo **ok**</code></pre>"
	if got := renderOne(t, md); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMarkdownToHTML_CodeBlockWithoutLanguage(t *testing.T) {
	if got, want := renderOne(t, "```\nls -la\n```"), "<pre>ls -la</pre>"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMarkdownToHTML_Lists(t *testing.T) {
	md := "- uno con `code`\n- dos\n  - anidado\n\n1. primero\n2. segundo\n\n- [x] hecho\n- [ ] pendiente"
	want := "• uno con <code>code</code>\n• dos\n    • anidado\n\n" +
		"1. primero\n2. segundo\n\n" +
		"• ☑ hecho\n• ☐ pendiente"
	if got := renderOne(t, md); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestMarkdownToHTML_Blockquote(t *testing.T) {
	if got, want := renderOne(t, "> **Nota:** cuidado"), "<blockquote><b>Nota:</b> cuidado</blockquote>"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMarkdownToHTML_Table(t *testing.T) {
	md := "| Recurso | Estado |\n|---|---|\n| bucket | ok |\n| vpc_main | **drift** |"
	want := "<pre>Recurso  | Estado\n---------+-------\nbucket   | ok\nvpc_main | drift</pre>"
	if got := renderOne(t, md); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestMarkdownToHTML_Empty(t *testing.T) {
	if chunks := markdownToHTML("  \n", msgutil.TelegramMaxMessageLength); len(chunks) != 0 {
		t.Errorf("expected no chunks, got %q", chunks)
	}
}

func TestMarkdownToHTML_SplitsBetweenBlocks(t *testing.T) {
	para := strings.Repeat("palabra ", 60) // 480 chars
	md := strings.Repeat("**"+para+"**\n\n", 20)

	chunks := markdownToHTML(md, 1000)
	if len(chunks) < 10 {
		t.Fatalf("expected the text to be split, got %d chunks", len(chunks))
	}
	for i, c := range chunks {
		if n := textLength(c); n > 1000 {
			t.Errorf("chunk %d has %d chars", i, n)
		}
		if strings.Count(c, "<b>") != strings.Count(c, "</b>") {
			t.Errorf("chunk %d has unbalanced tags: %q", i, c)
		}
	}
}

func TestMarkdownToHTML_SplitsLongCodeBlock(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("```yaml\n")
	for i := 0; i < 300; i++ {
		sb.WriteString("key_with_underscores: value <" + strings.Repeat("x", 20) + ">\n")
	}
	sb.WriteString("```")

	chunks := markdownToHTML(sb.String(), 1000)
	if len(chunks) < 2 {
		t.Fatalf("expected several chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if n := textLength(c); n > 1000 {
			t.Errorf("chunk %d has %d chars", i, n)
		}
		if !strings.HasPrefix(c, `<pre><code class="language-yaml">`) || !strings.HasSuffix(c, "</code></pre>") {
			t.Errorf("chunk %d is not a complete code block: %q", i, c)
		}
	}
}

// Telegram counts the limit after entity parsing, so the <pre><code> wrapper
// does not use any of it: a hard-cut line may fill the whole message.
func TestMarkdownToHTML_SplitsSingleLongCodeLine(t *testing.T) {
	line := strings.Repeat(`{"k":"a<b&c>d"},`, 700) // 11200 chars, no newlines
	chunks := markdownToHTML("```json\n"+line+"\n```", msgutil.TelegramMaxMessageLength)
	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks, got %d", len(chunks))
	}
	var got strings.Builder
	for i, c := range chunks {
		if n := textLength(c); n > msgutil.TelegramMaxMessageLength {
			t.Errorf("chunk %d has %d chars", i, n)
		}
		if !strings.HasPrefix(c, `<pre><code class="language-json">`) || !strings.HasSuffix(c, "</code></pre>") {
			t.Errorf("chunk %d is not a complete code block: %.60q", i, c)
		}
		got.WriteString(htmlToPlain(c))
	}
	if got.String() != line {
		t.Errorf("content lost across chunks: got %d chars, want %d", got.Len(), len(line))
	}
}

func TestMarkdownToHTML_CountsUTF16(t *testing.T) {
	md := strings.Repeat("🚀", 3000) // 3000 runes, 6000 UTF-16 code units
	for i, c := range markdownToHTML(md, msgutil.TelegramMaxMessageLength) {
		if n := textLength(c); n > msgutil.TelegramMaxMessageLength {
			t.Errorf("chunk %d has %d UTF-16 units", i, n)
		}
	}
}

func TestHTMLToPlain(t *testing.T) {
	in := `<b>Ojo</b>: <code>a &lt; b &amp;&amp; c</code> <a href="https://x.io">link</a>`
	if got, want := htmlToPlain(in), "Ojo: a < b && c link"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
