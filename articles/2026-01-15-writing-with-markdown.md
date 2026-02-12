---
title: Writing With Markdown
description: Why plain text is the best format for writing that lasts
slug: writing-with-markdown
date: 2026-01-15T10:00:00Z
tags: [writing, markdown, tools]
categories: [Essays]
draft: false
author: Gopher Mark
---

Every few years, a new writing tool appears that promises to revolutionize how we create content. Most fade away, taking your words with them. The ones that endure share a common trait: they build on plain text.

## The Portability Problem

Writing locked in a proprietary format is writing at risk. Export tools break. Companies shut down. APIs change. But a `.md` file you wrote ten years ago still opens in any text editor on any operating system.

Markdown gives you:

- **Portability** — your files work everywhere, forever
- **Version control** — track every change with git
- **Focus** — formatting gets out of the way
- **Flexibility** — convert to HTML, PDF, EPUB, or anything else

## Structure Without Overhead

Markdown's syntax maps directly to how we think about document structure:

```markdown
# This is a heading
## This is a subheading

A paragraph is just text with a blank line above it.

- Lists work naturally
- Like a bulleted outline

> Quotes are clear and obvious
```

There's no style panel to distract you. No font picker. No color wheel. Just your ideas, structured simply.

## Code Lives Alongside Prose

For technical writing, inline code like `fmt.Println("hello")` flows naturally in a sentence. Larger blocks get syntax highlighting:

```go
func main() {
    articles, err := loadArticles("./articles")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Loaded %d articles\n", len(articles))
}
```

This matters because the best technical writing interleaves explanation with working examples.

## The Editing Workflow

With files on disk, your editing workflow is whatever you want it to be:

1. Write in your favorite editor (VS Code, Vim, iA Writer, anything)
2. Preview with your blog engine's dev server
3. Commit when satisfied — git history is your revision log
4. Publish by pushing or restarting the server

No browser-based editor between you and your words. No save button. No cloud sync anxiety. Just files.

## Start Writing

The best format for writing is the one that disappears. Markdown disappears. Your ideas stay.
