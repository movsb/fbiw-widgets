// Package markdown renders GitHub-flavored Markdown directly into fbiw boxes.
package markdown

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/movsb/fbiw"
	"github.com/movsb/fbiw/widgets/list"
	"github.com/movsb/fbiw/widgets/table"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
)

// Renderer builds an fbiw tree without an intermediate HTML representation.
// Root is set after a successful goldmark Convert.
type Renderer struct {
	Document *fbiw.Document
	Root     fbiw.Box
}

func (*Renderer) AddOptions(...renderer.Option) {}

func (r *Renderer) Render(_ io.Writer, source []byte, node ast.Node) error {
	r.Root = nil
	if r.Document == nil {
		return fmt.Errorf("document is nil")
	}
	root, err := r.make(fbiw.NewBlock(r.Document))
	if err != nil {
		return err
	}
	if err := r.appendBlocks(root, node, source); err != nil {
		return err
	}
	r.Root = root
	return nil
}

// New returns a GFM parser configured to build boxes for doc.
func New(doc *fbiw.Document) goldmark.Markdown {
	return goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithRenderer(&Renderer{Document: doc}))
}

// Render parses source into a detached tree associated with doc. Append the
// returned box to an existing parent to include it in layout.
func Render(doc *fbiw.Document, source []byte) (fbiw.Box, error) {
	md := New(doc)
	if err := md.Convert(source, io.Discard); err != nil {
		return nil, err
	}
	return md.Renderer().(*Renderer).Root, nil
}

func (r *Renderer) make(box fbiw.Box, props ...string) (fbiw.Box, error) {
	for i := 0; i < len(props); i += 2 {
		if err := box.(fbiw.PropertySetter).SetProp(props[i], props[i+1]); err != nil {
			return nil, fmt.Errorf("%s %s: %w", box.GetTag(), props[i], err)
		}
	}
	return box, nil
}

func appendChild(parent, child fbiw.Box) error {
	if validator, ok := child.(interface{ ValidateChildren() error }); ok {
		if err := validator.ValidateChildren(); err != nil {
			return err
		}
	}
	if appender, ok := parent.(interface{ AppendChild(any) }); ok {
		appender.AppendChild(child)
	} else {
		parent.Base().AppendChild(child)
	}
	return nil
}

func appendText(parent fbiw.Box, value string) {
	parent.(interface{ AppendChild(any) }).AppendChild(value)
}

func (r *Renderer) appendBlocks(parent fbiw.Box, node ast.Node, source []byte) error {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		box, err := r.block(child, source)
		if err != nil {
			return err
		}
		if box != nil {
			if err := appendChild(parent, box); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Renderer) block(node ast.Node, source []byte) (fbiw.Box, error) {
	switch n := node.(type) {
	case *ast.Paragraph, *ast.TextBlock:
		if image, ok := node.FirstChild().(*ast.Image); ok && image.NextSibling() == nil {
			return r.make(fbiw.NewImage(r.Document), "src", string(image.Destination))
		}
		container, err := r.make(fbiw.NewBlock(r.Document), "padding", "0 0 8 0")
		if err != nil {
			return nil, err
		}
		text, err := r.text(node, source)
		if err != nil {
			return nil, err
		}
		return container, appendChild(container, text)
	case *ast.Heading:
		container, err := r.make(fbiw.NewBlock(r.Document), "padding", "12 0 6 0")
		if err != nil {
			return nil, err
		}
		text, err := r.text(node, source, "bold", "true", "font-size", strconv.Itoa(max(22, 38-(n.Level-1)*3)))
		if err != nil {
			return nil, err
		}
		return container, appendChild(container, text)
	case *ast.List:
		var widget *list.List
		if n.IsOrdered() {
			widget = list.NewOrderedList(r.Document)
		} else {
			widget = list.NewUnorderedList(r.Document)
		}
		listBox, err := r.make(widget)
		if err != nil {
			return nil, err
		}
		if n.IsOrdered() {
			if err := widget.SetProp("start", strconv.Itoa(n.Start)); err != nil {
				return nil, err
			}
		}
		return listBox, r.appendBlocks(listBox, node, source)
	case *ast.ListItem:
		item, err := r.make(list.NewItem(r.Document))
		if err != nil {
			return nil, err
		}
		return item, r.appendBlocks(item, node, source)
	case *extast.Table:
		tableBox, err := r.make(table.NewTable(r.Document), "border-width", "1", "border-color", "#777777")
		if err != nil {
			return nil, err
		}
		return tableBox, r.appendBlocks(tableBox, node, source)
	case *extast.TableHeader, *extast.TableRow:
		row, err := r.make(table.NewTableRow(r.Document))
		if err != nil {
			return nil, err
		}
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			var cellBox *table.TableCell
			if _, ok := n.(*extast.TableHeader); ok {
				cellBox = table.NewTableHeaderCell(r.Document)
			} else {
				cellBox = table.NewTableCell(r.Document)
			}
			cell, err := r.make(cellBox, "padding", "5 8")
			if err != nil {
				return nil, err
			}
			text, err := r.text(child, source)
			if err != nil {
				return nil, err
			}
			if err := appendChild(cell, text); err != nil {
				return nil, err
			}
			if err := appendChild(row, cell); err != nil {
				return nil, err
			}
		}
		return row, nil
	case *ast.Blockquote:
		quote, err := r.make(fbiw.NewBlock(r.Document), "padding", "4 0 4 16", "border-width", "2", "border-color", "#777777")
		if err != nil {
			return nil, err
		}
		return quote, r.appendBlocks(quote, node, source)
	case *ast.FencedCodeBlock, *ast.CodeBlock:
		code, err := r.make(fbiw.NewBlock(r.Document), "padding", "8", "background-color", "#242830")
		if err != nil {
			return nil, err
		}
		for i := 0; i < node.Lines().Len(); i++ {
			line := node.Lines().At(i)
			text, err := r.make(fbiw.NewText(r.Document))
			if err != nil {
				return nil, err
			}
			appendText(text, strings.TrimSuffix(string(line.Value(source)), "\n"))
			if err := appendChild(code, text); err != nil {
				return nil, err
			}
		}
		return code, nil
	case *ast.ThematicBreak:
		return r.make(fbiw.NewBlock(r.Document), "height", "1", "background-color", "#777777")
	case *ast.HTMLBlock:
		// Raw HTML is not interpreted as fbiw components.
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported Markdown block: %s", node.Kind())
	}
}

func (r *Renderer) text(node ast.Node, source []byte, props ...string) (fbiw.Box, error) {
	text, err := r.make(fbiw.NewText(r.Document), props...)
	if err != nil {
		return nil, err
	}
	return text, r.inlines(text, node, source)
}

func (r *Renderer) inlines(parent fbiw.Box, node ast.Node, source []byte) error {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			appendText(parent, string(n.Value(source)))
			if n.HardLineBreak() || n.SoftLineBreak() {
				appendText(parent, " ")
			}
		case *ast.String:
			appendText(parent, string(n.Value))
		case *ast.Emphasis, *ast.CodeSpan:
			var segment fbiw.Box = fbiw.NewCodeText(r.Document)
			if emphasis, ok := n.(*ast.Emphasis); ok {
				segment = fbiw.NewItalicText(r.Document)
				if emphasis.Level >= 2 {
					segment = fbiw.NewBoldText(r.Document)
				}
			}
			segment, err := r.make(segment)
			if err != nil {
				return err
			}
			if err := r.inlines(segment, child, source); err != nil {
				return err
			}
			if err := appendChild(parent, segment); err != nil {
				return err
			}
		case *ast.AutoLink:
			appendText(parent, string(n.Label(source)))
		case *ast.Link, *ast.Image, *extast.Strikethrough:
			// Links have no activation yet; inline images show their alt text.
			if err := r.inlines(parent, child, source); err != nil {
				return err
			}
		case *extast.TaskCheckBox:
			marker := "☐ "
			if n.IsChecked {
				marker = "☑ "
			}
			appendText(parent, marker)
		case *ast.RawHTML:
			// Never execute source HTML as components.
		default:
			if err := r.inlines(parent, child, source); err != nil {
				return err
			}
		}
	}
	return nil
}
