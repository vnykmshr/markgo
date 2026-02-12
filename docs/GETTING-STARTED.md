# Getting Started

> Zero to running blog in under 5 minutes.

---

## Install

**Download a release** from [GitHub Releases](https://github.com/vnykmshr/markgo/releases), or build from source:

```bash
git clone https://github.com/vnykmshr/markgo.git && cd markgo
make build
```

The binary lands at `./build/markgo`.

## Initialize

```bash
markgo init --quick
```

This creates a `.env` config, an `articles/` directory with a sample post, and the `web/` directory with templates and static assets.

## Run

```bash
markgo serve
```

Visit http://localhost:3000. Your blog is live.

---

## Write Something

MarkGo has three content types. You never pick one — the system infers it from what you write.

### From the command line

```bash
# Article — has a title, intended for long-form
markgo new --title "How I Built This" --tags "golang,blogging"

# Thought — no title needed, short-form
markgo new --type thought

# Link — sharing a URL with commentary
markgo new --type link
```

Articles are markdown files in `articles/` with YAML frontmatter:

```markdown
---
title: "How I Built This"
description: "A walkthrough of the architecture"
tags: ["golang", "blogging"]
category: "engineering"
date: 2026-02-12T00:00:00Z
---

Your content here, in standard markdown.
```

### From the browser

If you configure admin credentials in `.env`:

```bash
ADMIN_USERNAME=you
ADMIN_PASSWORD=something-strong
```

Then restart the server. You'll see a Compose link in the nav and a floating action button (FAB) on mobile. The compose form puts the cursor in the content field — not the title, not a category picker. Write first, categorize never.

**Quick capture**: Tap the FAB, type a thought, hit Publish. Under 5 seconds.

**Full compose**: Navigate to `/compose` for the full form with title, tags, markdown preview, and image upload.

**Edit**: Any published post can be edited at `/compose/edit/:slug`.

### Content type inference

You don't choose a type. MarkGo figures it out:

| What you write | What it becomes |
|---|---|
| No title, under 100 words | Thought |
| Has a `link_url` in frontmatter | Link |
| Everything else | Article |

You can override this by setting `type: thought`, `type: link`, or `type: article` in the frontmatter.

---

## Configure

Edit `.env`. The essential settings:

```bash
# Your blog
BLOG_TITLE=My Blog
BLOG_AUTHOR=Your Name
BASE_URL=http://localhost:3000

# For production
ENVIRONMENT=production
BASE_URL=https://yourdomain.com
```

See [configuration.md](configuration.md) for every option.

---

## Development

```bash
make dev      # Live reload server at :3000 (requires air)
make build    # Build binary
make test     # Run tests
make lint     # Run linter
```

---

## Deploy

**Docker:**
```bash
docker compose up -d
```

**Manual:**
```bash
make build
scp build/markgo server:/usr/local/bin/
```

**Static export** (GitHub Pages, Netlify, Vercel):
```bash
markgo export --output ./public --base-url https://yourdomain.com
```

See [deployment.md](deployment.md) for reverse proxy setup, systemd, and production configuration.

---

## Features You Get Out of the Box

- **SPA navigation** — Instant page transitions, no full reloads
- **PWA** — Installable, works offline, caches pages for offline reading
- **Quick capture** — FAB on mobile, Cmd/Ctrl+N on desktop
- **Offline compose** — Write when disconnected, auto-syncs when back online
- **Search** — Full-text search across all content
- **Feeds** — RSS (`/feed.xml`), JSON Feed (`/feed.json`), sitemap (`/sitemap.xml`)
- **SEO** — Open Graph, Twitter Cards, Schema.org, canonical URLs
- **Themes** — Light/Dark/Auto mode, 5 color presets, 3 style themes
- **Comments** — Optional Giscus integration
- **Contact form** — SMTP email delivery with rate limiting
- **Admin panel** — Stats, draft management, cache controls

All of this works without JavaScript for core reading. JS enhances, it doesn't gate.

---

## Next Steps

- [Configuration](configuration.md) — All environment variables
- [Architecture](architecture.md) — How the system works
- [API](api.md) — Every HTTP endpoint
- [Deployment](deployment.md) — Production setup
- [Design Language](design.md) — The principles behind the decisions
