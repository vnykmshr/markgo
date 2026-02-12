# MarkGo Design Language

> A living document. Updated as the product evolves.
> Last revised: 2026-02-12

---

## Positioning

MarkGo is the only blog engine where you write first and categorize never.

Type two sentences without a title and it becomes a thought. Paste a URL with commentary and it becomes a link. Write something long with a title and it becomes an article. You never choose a post type. You just write. Everything lands in one chronological feed, the way your mind actually works — not in neat buckets a CMS imposed on you.

Most blog engines are either static site generators (Hugo, Jekyll — files on disk, but no live server) or hosted platforms (Ghost, Substack — live server, but no file ownership). MarkGo is neither. It reads markdown files from your filesystem _and_ serves them live with search, feeds, and a web compose form. You can draft in vim on your laptop, then publish a quick thought from your phone on the bus. Both workflows write the same markdown files to the same directory.

The mental model: a notebook that happens to have a URL. No database. No build step. No deploy pipeline more complex than copying files.

---

## The Feeling

MarkGo should feel like opening a notebook to a blank page — not like logging into a system.

The compose form puts the cursor in the content field. Not the title. Not the category picker. The content. Everything else is optional, and the system infers what it can. The distance between "I have something to say" and "it's said" should be as close to zero as the web allows.

For the reader: the feed is one person's unfiltered stream of thinking. Thoughts, links, and essays mixed together, the way conversation works. No algorithmic reordering. No "you might also like." Just: here is what this person has been thinking about, in the order they thought it.

---

## Design Principles

Five beliefs that guide every decision. A principle that nobody would disagree with is not doing work — these are meant to create productive tension.

### 1. Content-first, chrome-last

The feed _is_ the homepage. Navigation is two text links — "Writing" and "About" — plus a search icon and RSS button. When admin credentials are configured, a third link ("Compose") appears. No sidebar. No hero image. No "featured posts" carousel. The content appears immediately, and everything else earns its pixel.

**Say no to:** dashboard widgets, sidebar recommendations, header banners, anything that pushes the first piece of content below the fold.

### 2. Thought-sized to essay-length

Publishing frequency shouldn't require ceremony. A two-sentence observation and a 3,000-word essay are equally valid acts of publishing. The content model treats thoughts, links, and articles as first-class citizens — different cards in the same stream, not second-class posts relegated to a "microblog" section.

**Say no to:** separate "microblog" pages, minimum word counts, mandatory titles, post-type selection dialogs.

### 3. Mobile is the default, not the fallback

CSS starts at 320px and works up. Not 1024px scaled down. Interactive elements maintain a minimum 40px touch target (verified: `.nav-action-btn` and `.theme-toggle` are 40x40px). If an interaction requires a hover state for critical information, it is a bug.

**Say no to:** hover-only tooltips, right-click context menus, drag-and-drop as the only way to reorder, layouts that require horizontal scrolling on phone screens.

### 4. Local files, global reach

Articles are markdown files with YAML frontmatter. They live on your filesystem. You can edit them in any text editor, version them with git, back them up however you like. MarkGo reads the directory and serves what it finds. No database migration. No import/export dance. The files _are_ the source of truth. The compose form (`/compose`) is a convenience layer — it writes markdown files to disk, not to a database.

**Say no to:** database-backed content storage, proprietary formats, features that bypass the filesystem.

### 5. Degrade gracefully, depend on nothing

The core reading experience must work if every CDN goes down. Fonts degrade to system stacks. Syntax highlighting degrades to unstyled `<pre>` blocks. Filter pills are `<a>` tags, not JavaScript event handlers. Forms use native POST actions. Without JavaScript, you get system-preference theming, a visible nav, and fully functional browsing. Nothing breaks — features degrade.

**Say no to:** JavaScript-required interactions for core reading, client-side rendering, features that fail silently when a third-party service is unavailable.

All fonts and highlight.js are self-hosted. The binary embeds all web assets — no external CDN dependencies at runtime.

---

## Voice & Tone

How MarkGo speaks in its interface:

| We say | Not |
|--------|-----|
| **Compose** | Create Post |
| **Writing** | Blog |
| **What's on your mind?** | Enter content |
| **Written in Markdown, published with MarkGo** | Powered by MarkGo |
| **Publish** | Submit |
| **Save as draft** | Mark as unpublished |
| **No posts yet** | No content found |
| **No results found** | We couldn't find any articles |

The voice is **conversational and direct**. It sounds like a person, not a product. Labels are verbs when possible ("Compose", "Subscribe", "Search"), not nouns ("Post Creator", "Feed Reader"). Hints guide without lecturing: "optional — leave empty for thoughts" tells you the consequence of your choice, not just the constraint.

**Error and empty states:** explain what happened and what to do. Never blame the user. Never say "invalid." Use second-person address ("Nothing matched your search"), not institutional "we" ("We couldn't find anything").

**Tone across contexts:**
- Compose form: intimate, encouraging ("What's on your mind?")
- Navigation and labels: minimal, functional ("Writing", "About", "Search")
- Empty states: honest, brief ("No posts yet")
- Errors: calm, helpful ("Page not found. The page you're looking for doesn't exist or has been moved.")

---

## The Three Streams

MarkGo's content model has three types, inferred automatically from what you write. You never pick a type from a dropdown — the system figures it out from the shape of your content. This is MarkGo's most distinctive design decision.

**Thoughts** — No title, under 100 words. A fleeting idea, a reaction, a note-to-self that's worth sharing. Displayed as a simple text card with a left accent stripe in `--color-primary`. No "read more" link — the whole thought is right there. Shows relative time ("2 hours ago") and tags.

**Links** — A URL you're sharing, with optional commentary. Displayed with the domain extracted (e.g., "github.com") and a visit arrow (→). Links to both the MarkGo article page (title) and the external URL (domain). Shows relative time and tags.

**Articles** — The traditional blog post. Title, description or excerpt, absolute date ("Jan 2, 2006"), reading time ("5 min read"), and tags. The long-form piece you drafted over days.

The visual density increases with content commitment: thoughts show only text and time, links add a title and domain, articles add description and reading time. This progressive density is intentional — lighter content gets lighter chrome.

The inference rules are intentionally simple (see `internal/services/article/inference.go`):
- Explicit `type` in frontmatter always wins
- Has `link_url` → link
- No title and under 100 words → thought
- Everything else → article

This means you never _have_ to think about types. Write naturally. MarkGo figures it out. But if you want control, the frontmatter `type` field overrides inference.

All three types live in the same feed, filterable by type via server-rendered `<a>` tag pills with query parameters (`/?type=thought`). No JavaScript required to browse by type.

---

## Visual Language

The design tokens live in `web/static/css/main.css`. This section explains _why_ they exist, not what they are.

### Typography

**Inter** — A geometric sans-serif that's neutral enough to disappear behind content. It doesn't impose a personality. It reads well at small sizes on mobile screens. Self-hosted with system-stack fallback: `-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif`.

**Fira Code** — Monospace with ligatures for code blocks. Chosen for readability in technical content.

### Surfaces and Shape

**Border radius: `--radius-lg` (0.5rem)** — Friendly, not bubbly. Cards and inputs feel approachable without looking childish. Smaller radii (`--radius-base` 0.25rem, `--radius-md` 0.375rem) for inline elements. `--radius-full` for pills and badges.

**Shadows: subtle, layered** — Cards lift slightly on the z-axis. They don't jump off the page. Shadow tokens progress from `--shadow-sm` (barely there) to `--shadow-xl` (modal-level), but most UI uses `--shadow-sm` and `--shadow-base`. Elevation is information architecture, not decoration.

### Motion

**`--transition-fast` (150ms)** — The default for hover states, focus rings, and button interactions. Responsive, not bouncy. The UI acknowledges your action without making you wait.

**`--transition-base` (250ms)** — For component-level state changes: card hover lift, accordion expand.

**`--transition-slow` (350ms)** — Reserved for page-level transitions. Rarely used.

`@media (prefers-reduced-motion: reduce)` disables all animations globally.

### Spacing

Spacing tokens start at 4px (`--spacing-1`) and scale through 80px (`--spacing-20`). The feed stream uses enough vertical gap that each card reads as an independent thought, not a row in a table. Generous whitespace is a feature, not wasted space.

### Theming: Two Independent Axes

MarkGo has two customization layers that work independently:

**Color presets** (`data-color-theme`) — Swap the `--color-primary` family. Five ship out of the box: default, ocean, forest, sunset, berry. Defined as `[data-color-theme="..."]` selectors in `main.css`. These are cosmetic — they change accent colors only.

**Style themes** (`Blog.Style`) — Load an additional CSS file from `web/static/css/themes/` that overrides `--theme-*` variables for typography, spacing, and broader visual personality. Ships with `minimal`. The system is extensible — add a CSS file to `themes/` and set `BLOG_STYLE` to its name. Style themes are structural — they can change the entire feel of the site.

New themes must only set `--theme-*` variables — never override component selectors directly.

**Dark mode** — Follows system preference automatically via `@media (prefers-color-scheme: dark)`. Manual toggle stores preference in localStorage as `data-theme`. Dual-selector pattern in `main.css`: system preference applies to `:root:not([data-theme="light"])`, manual toggle applies to `[data-theme="dark"]`.

---

## Interaction Principles

**Server-rendered first** — The page works with JavaScript disabled. Forms use native POST actions. Filter pills are `<a>` tags with query parameters. The server does the work; the client enhances.

**SPA navigation** — With JavaScript, the app shell router intercepts link clicks, fetches full HTML, swaps `<main>`, and pushes history state. The server still renders complete pages — the client just avoids full reloads. Prefetch on hover (65ms delay) makes transitions feel instant. A CSS-only progress bar shows during navigation. Falls back to full page loads without JS.

**Progressive enhancement** — Theme toggle, bottom navigation, compose sheet, and code-block copy buttons are JavaScript features. Without JS, you get system-preference theming, a nav bar with links, and fully functional browsing. Nothing breaks — features degrade.

**Mobile-native UX** — Bottom navigation (5-item tab bar) replaces the header nav on mobile. Full-screen search overlay slides up from the bottom nav. Visual viewport handlers reposition overlays above the iOS keyboard. Dynamic `theme-color` meta tag matches the page background. The app should feel like it belongs on the home screen, not in a browser tab.

**Quick capture** — A floating action button (FAB) triggers a compose sheet overlay. On mobile: bottom sheet sliding up. On desktop: centered modal. Type content, hit Publish, done. Auto-save drafts to localStorage with 2-second debounce and recovery on re-open. Keyboard shortcut: Cmd/Ctrl+N. The goal: under 5 seconds from thought to published.

**Offline and installable** — Service Worker caches pages for offline reading and queues compose submissions in IndexedDB for sync on reconnect. The PWA is installable with a dynamic web manifest generated from config. Network-only routes (admin, compose, login, feeds) are never cached.

**FOUC prevention** — An inline `<script>` in `<head>` reads the theme preference from localStorage before the first paint. Wrapped in try/catch; fails silently.

**Pagination, not infinite scroll** — "Page 1 of 3" with Newer/Older links. Your position in the feed is stable and bookmarkable. Attention is a finite resource; the UI respects it.

**Compose and admin** — The compose form at `/compose` and the admin dashboard at `/admin` are gated behind session authentication configured in `.env`. These exist to enable publishing from a phone or tablet — the "train with one thumb" scenario. They are optional; you can ignore them entirely and manage articles as files. The compose form writes markdown files to disk — it is a convenience layer over the filesystem, not a replacement for it.

---

## Architecture Decisions

The Interaction Principles describe _what_ the UI does. This section explains _why_ those choices were made instead of the obvious alternatives.

### App Shell: Turbo Drive, Not a Framework

MarkGo uses a Turbo Drive-style SPA router — intercept clicks, fetch full HTML, swap `<main>`, push history. The server still renders complete pages. The client just avoids full reloads.

**Why not React/Vue/Svelte?** MarkGo already has a Go template engine that renders every page. Adding a JS framework would mean rendering content twice: once on the server (for no-JS, SEO, first paint) and once on the client (for SPA transitions). That's two codebases for the same UI. The Turbo Drive pattern gets SPA-feel navigation with zero client-side templating — the server is the single source of truth.

**Why not htmx?** htmx is a dependency. MarkGo's router is 335 lines of vanilla JS with zero dependencies. htmx would also require partial HTML responses — MarkGo serves complete pages, which means every URL works as a direct link, a browser bookmark, and a shared URL without special server handling.

**Trade-off:** No client-side state management. Every "page" is a fresh server render. This is fine for a blog — there's no complex interactive state to manage.

### Mobile: Bottom Nav, Not Hamburger

Mobile navigation uses a 5-item bottom tab bar instead of a hamburger menu.

**Why?** Hamburger menus hide navigation behind a tap — users have to remember what's available. Bottom tabs show everything at a glance, are reachable with one thumb, and match the native app conventions users already know (iOS tab bar, Android bottom nav). Studies consistently show bottom tabs get more engagement than hamburger menus.

**Why 5 items?** Home, Writing, Compose (+), Search, About. This covers the primary user journeys. Apple's HIG recommends 3-5 items. More than 5 becomes cramped on narrow screens.

**Trade-off:** The compose button is always visible even for unauthenticated visitors (though it triggers a login flow). This is intentional — it signals that MarkGo is a tool for writing, not just reading.

### PWA: Offline Reading + Compose Queue

The Service Worker implements 3-tier caching: precache (offline fallback), stale-while-revalidate (static assets), and network-first (HTML pages). Compose submissions queue in IndexedDB when offline and auto-sync on reconnect.

**Why PWA instead of native apps?** MarkGo is a single-developer project. PWA gives install-to-homescreen, offline support, and full-screen mode without maintaining iOS and Android codebases. The trade-off: no push notifications, no background sync on iOS (compose queue only drains when the app is foregrounded).

**Why IndexedDB for compose queue, not localStorage?** localStorage is synchronous, has a 5-10MB limit, and can be cleared by the browser during storage pressure. IndexedDB is async, has larger quotas, and is explicitly designed for structured data persistence. For queued form submissions that the user expects to survive app restarts, IndexedDB is the right tool.

**Network-only routes:** Admin, compose, login, logout, feeds, and API endpoints are never cached. Stale admin data or cached auth state would create confusing bugs.

---

## Web Standards

Structural HTML and SEO conventions used across all templates.

### Heading Hierarchy

One `<h1>` per page. The site title in the header is a `<span>`, not a heading — it's branding, not content structure. Each page template defines its own `<h1>` (the feed page uses an `sr-only` h1 since the feed has no visible heading). Card titles within listings are `<h2>`, never `<h3>` — no heading level skips.

### Semantic Containers

Content streams use `<section>` with `aria-label` instead of generic `<div>`. Contact information uses `<address>`. Navigation regions use `<nav>` with descriptive `aria-label` values. Breadcrumbs use `<nav aria-label="Breadcrumb">` with an ordered list (`<ol>`) and `aria-current="page"` on the current item.

### Canonical URLs

Every public page sets `canonicalPath` in its handler, rendered as `<link rel="canonical" href="{{ baseURL }}{{ canonicalPath }}">`. This covers: `/`, `/writing`, `/writing/{slug}`, `/tags`, `/tags/{tag}`, `/categories`, `/categories/{category}`, `/search`, `/about`.

### Structured Data (JSON-LD)

| Page Type | Schema | Source |
|-----------|--------|--------|
| All pages | WebSite (with SearchAction) | SEO service via `enhanceTemplateDataWithSEO` |
| All pages | BreadcrumbList | SEO service via `seo_helper.go` |
| Article | BlogPosting (headline, author, publisher+logo, dates, wordCount) | Inline in `article.html` |
| Article | Article (from SEO service) | SEO service via `articleSchema` |
| Tag page | CollectionPage + ItemList | Handler-built in `taxonomy_handler.go` |
| Category page | CollectionPage + ItemList | Handler-built in `taxonomy_handler.go` |

### External Links

All `target="_blank"` links include `rel="noopener"`. The SPA router only intercepts same-origin links.

---

## Anti-Patterns

Things MarkGo deliberately does not do. Each one names a trade-off, not just an absence.

**No rich text editor** — The compose form is a raw markdown textarea. Ghost, Substack, Medium, and Micro.blog all have rich text editors. MarkGo deliberately chose not to build one. The trade-off: MarkGo is not for people who don't know what `**bold**` means. The benefit: no editor complexity, no format lock-in, files are portable.

**No explicit post-type selection** — You never choose "thought" vs "link" vs "article." The inference engine figures it out from what you wrote. Most platforms with multiple post types (Tumblr, Micro.blog) require you to pick one up front. The trade-off: you can't override inference from the compose form (only from frontmatter). The benefit: zero friction between "I have a thought" and "it's published."

**No build step** — CSS and JS are vanilla. No webpack, no Tailwind compilation, no `npm run build`. The trade-off: no tree-shaking, no CSS modules, no TypeScript. The benefit: `git clone && make dev` is all you need.

**No engagement metrics** — No like counts, no view counts, no "trending" badges. Writing is its own reward. The trade-off: you will never know if anyone read what you wrote from MarkGo itself. The benefit: the absence of metrics removes the anxiety of performance.

**No first-party tracking** — No analytics scripts, no fingerprinting, no first-party tracking cookies. The trade-off: no usage data to inform design decisions. The benefit: no cookie consent banner needed. All assets are self-hosted — zero external requests.

**No dark patterns** — No newsletter popup on first visit. No "subscribe before you read" gate. No exit-intent modals. Content is freely accessible the moment you arrive.

---

## The Litmus Test

Three questions for every design decision:

1. **Does this serve the writer or the platform?**
   If a feature exists to grow the platform rather than help the writer write, it doesn't belong here.

2. **Would this work on a train with one thumb?**
   If an interaction requires precise mouse targeting or a wide viewport, redesign it.

3. **Does this need JavaScript, or is HTML enough?**
   Start with a server-rendered HTML solution. Add JavaScript only when the HTML version is genuinely worse, not just less flashy.

---

## Maintaining This Document

Update this document when:
- A new content type is added to the inference rules
- A new principle is needed to resolve a recurring design disagreement
- A claim in this document no longer matches the codebase
- An anti-pattern is intentionally violated (remove it or explain the exception)

Every update gets a conventional commit: `docs(design): <what changed and why>`.

---

_This document is the compass, not the map. When it conflicts with user needs observed in practice, update the document — don't ignore the observation._
