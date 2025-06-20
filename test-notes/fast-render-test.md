# Simple Markdown Renderer Test

This tests our fast, lightweight markdown renderer.

## Features

The renderer supports:

- **Bold text** for emphasis
- *Italic text* for subtle emphasis  
- `Inline code` for technical terms
- Bullet points like this one

### Code Blocks

```go
func fastRender(content string) string {
    // Much faster than glamour!
    return renderSimpleMarkdown(content)
}
```

> Blockquotes are also supported
> They can span multiple lines

---

That's a horizontal rule above!

## Performance

This renderer is:
- Lightning fast
- No external dependencies (just lipgloss)
- Perfect for instant previews
