# Markdown

[fbiw-widgets](../README.md) 中的 Markdown 渲染组件。使用
[goldmark](https://github.com/yuin/goldmark) 的 GFM 扩展解析 Markdown，
并用自定义 goldmark 渲染器调用 fbiw 与各组件的构造函数，直接构造组件树，
不经过 HTML 字符串或二次解析。

```go
import (
    "github.com/movsb/fbiw-widgets/markdown"
)

tree, err := markdown.Render(doc, []byte("# 标题\n\n- 一项\n- 二项"))
if err != nil {
    return err
}
container.AppendChild(tree)
```

`Render` 返回关联到 `doc` 的独立根盒子。调用者负责把它加入已有文档，
例如把它放进 `<scroll>` 中。包内导入列表和表格组件，无需调用方另行注册标签。

目前支持标题、段落、粗体、斜体、行内代码、代码块、引用、分隔线、
有序与无序列表、GFM 表格、任务列表，以及单独成段的图片。原始 HTML 不会执行。

当前边界：链接只显示文字、不能激活；删除线只显示文字；行内图片显示替代文字；
代码块尚未保证完整的等宽字体与空白排版语义。根目录 `go.mod` 中的相邻目录 `replace` 用于本地联调，
准备独立发布时应改为已发布的 fbiw 版本。

运行交互示例：在仓库根目录执行 `go run ./markdown/demo`。示例内容位于
[`demo/sample.md`](demo/sample.md)，方向键滚动，L1/R1 调整字号。
