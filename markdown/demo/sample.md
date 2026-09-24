# A small GFM document

This page is **Markdown**, rendered as an fbiw box tree. It also has *emphasis*, `inline code`, and a [link label](https://example.com).

## Tasks

- [x] Parse GFM with goldmark
- [x] Build text, lists, and tables directly
- [ ] Add interactive links later

1. Ordered items work too.
2. They can include **styled text**.
   - A nested bullet.

## Auto-sized table

| Feature | Current behavior | Note |
| --- | --- | --- |
| Lists | Nested and wrapped | Ordered and unordered |
| Tables | Content-sized columns | Collapsed grid |
| Links | Text only | No activation yet |

> A quote is a normal fbiw block with a border and padding. Resize the window or change font size to see the content reflow.

## Code

```go
tree, err := markdown.Render(doc, source)
if err != nil {
    return err
}
container.AppendChild(tree)
```

---

Raw HTML is intentionally ignored rather than executed as components.
