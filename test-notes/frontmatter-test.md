---
title: YAML Frontmatter Test
date: 2025-06-19
tags:
  - test
  - frontmatter
  - preview
author: Test User
---

# This Should Be The First Visible Line

The YAML frontmatter above should not appear in the preview.

## Regular Content

This is the actual content of the note that users want to see.

- The frontmatter metadata is hidden
- Only the markdown content is shown
- Much cleaner preview experience

### Benefits

1. **Cleaner previews** - No metadata clutter
2. **Faster scanning** - Jump right to content
3. **Better UX** - See what matters

> "The best interface is no interface" - Golden Krishna

```javascript
// Code should still work fine
const preview = renderMarkdown(content);
console.log("Frontmatter removed!");
```

Perfect for notes that use frontmatter for organization but don't need it displayed.
