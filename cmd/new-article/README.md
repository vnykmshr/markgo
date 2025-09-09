# new-article CLI Tool

A simple command-line tool for creating markdown blog articles with YAML frontmatter.

## Installation

```bash
# Build the tool
make new-article

# The binary will be available at ./build/new-article
```

## Usage

### Interactive Mode (Default)
Run without arguments to be prompted for each field:

```bash
./build/new-article
```

### Command-Line Mode
Specify arguments directly:

```bash
./build/new-article --title "My Article" --tags "golang,web" --category "programming" --draft=false
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `--title` | Article title | "Untitled Article" |
| `--description` | Brief description | "" |
| `--tags` | Comma-separated tags | "general" |
| `--category` | Article category | "uncategorized" |
| `--author` | Author name | Current OS username |
| `--draft` | Mark as draft | `true` |
| `--featured` | Mark as featured | `false` |
| `--help` | Show help | - |

## Examples

```bash
# Interactive mode
./build/new-article

# Quick article
./build/new-article --title "Hello World"

# Complete article
./build/new-article --title "Go Tutorial" --description "Learn Go" --tags "golang,tutorial" --category "programming" --draft=false --featured=true
```

## Output

Creates markdown files in `articles/` directory with format:
- Filename: slugified title (e.g., "Hello World" → `hello-world.md`)
- YAML frontmatter with metadata
- Basic markdown template

## Features

- 🚀 Interactive prompts with defaults
- 📁 Automatic filename generation
- 🛡️ Prevents overwriting existing files
- 📝 Structured markdown template
- 🎯 Zero external dependencies