package markdown

import (
	"os"
	"testing"

	"github.com/movsb/fbiw"
)

func TestGFMTree(t *testing.T) {
	source := []byte("# Title\n\nSome **bold** and `code`.\n\n- [x] Done\n- [ ] Todo\n\n| A | B |\n|---|---|\n| 1 | 2 |\n")
	root, err := Render(&fbiw.Document{}, source)
	if err != nil {
		t.Fatal(err)
	}
	if len(root.Children()) != 7 || root.Children()[4].GetTag() != `ul` || root.Children()[6].GetTag() != `table` {
		t.Fatalf("unexpected block tree: %v", root.Children())
	}
	for _, index := range []int{1, 3, 5} {
		if root.Children()[index].GetTag() != `spacer` {
			t.Fatalf("missing top-level spacer at %d: %v", index, root.Children())
		}
	}
	paragraph := root.Children()[2].Children()[0]
	if paragraph.GetTag() != `text` || len(paragraph.Children()) != 2 || paragraph.Children()[0].GetTag() != `b` || paragraph.Children()[1].GetTag() != `code` {
		t.Fatalf("unexpected inline tree: %v", paragraph.Children())
	}
}

func TestRenderBoxTree(t *testing.T) {
	doc := &fbiw.Document{}
	root, err := Render(doc, []byte("# Title\n\n- one\n- two\n\n| A |\n|---|\n| B |\n"))
	if err != nil {
		t.Fatal(err)
	}
	if root.GetTag() != `block` || len(root.Children()) != 5 {
		t.Fatalf("root = %s, children = %d", root.GetTag(), len(root.Children()))
	}
	if root.Children()[2].GetTag() != `ul` || root.Children()[4].GetTag() != `table` {
		t.Fatalf("children = %s, %s", root.Children()[2].GetTag(), root.Children()[4].GetTag())
	}
	if len(root.Children()[2].Children()) != 2 {
		t.Fatalf("list items = %v", root.Children()[2].Children())
	}
}

func TestTopLevelSpacingFollowsFontSize(t *testing.T) {
	root, err := Render(&fbiw.Document{}, []byte("first\n\nsecond"))
	if err != nil {
		t.Fatal(err)
	}
	if len(root.Children()) != 3 {
		t.Fatalf("children = %v", root.Children())
	}
	spacer, ok := root.Children()[1].(*_EmSpacer)
	if !ok {
		t.Fatalf("spacing box = %T", root.Children()[1])
	}
	for _, size := range []int{16, 24} {
		spacer.GetComputedStyles().FontSize = fbiw.NumberLength(size)
		spacer.Calc(100, 100, fbiw.Constraints{})
		if got := spacer.GetLayoutBox().Height; got != size {
			t.Fatalf("font size %d: spacing = %d", size, got)
		}
	}
}

func TestRawHTMLNotInterpreted(t *testing.T) {
	root, err := Render(&fbiw.Document{}, []byte("<button>unsafe</button>\n\nhello <img src=evil> world"))
	if err != nil {
		t.Fatal(err)
	}
	for _, child := range root.Children() {
		if child.GetTag() == `button` || child.GetTag() == `img` {
			t.Fatalf("raw HTML became component: %s", child.GetTag())
		}
	}
}

func TestStandaloneImage(t *testing.T) {
	root, err := Render(&fbiw.Document{}, []byte("![alt](picture.png)"))
	if err != nil {
		t.Fatal(err)
	}
	if len(root.Children()) != 1 || root.Children()[0].GetTag() != `img` {
		t.Fatalf("image = %v", root.Children())
	}
}

func TestCodeBlockKeepsRows(t *testing.T) {
	root, err := Render(&fbiw.Document{}, []byte("```go\n  one\n\n  two\n```\n"))
	if err != nil {
		t.Fatal(err)
	}
	code := root.Children()[0]
	if len(code.Children()) != 1 || code.Children()[0].GetTag() != `text` {
		t.Fatalf("code children = %v", code.Children())
	}
	if got := code.Children()[0].(*fbiw.Text).GetText(); got != "  one\n\n  two" {
		t.Fatalf("code text = %q", got)
	}
}

func TestInlineWhitespaceIsPreserved(t *testing.T) {
	root, err := Render(&fbiw.Document{}, []byte("before **bold** after"))
	if err != nil {
		t.Fatal(err)
	}
	text := root.Children()[0].Children()[0].(*fbiw.Text)
	if got := text.GetText(); got != "before bold after" {
		t.Fatalf("text = %q", got)
	}
}

func TestDemoSample(t *testing.T) {
	source, err := os.ReadFile("demo/sample.md")
	if err != nil {
		t.Fatal(err)
	}
	root, err := Render(&fbiw.Document{}, source)
	if err != nil {
		t.Fatal(err)
	}
	if len(root.Children()) < 6 {
		t.Fatalf("demo rendered only %d blocks", len(root.Children()))
	}
}
