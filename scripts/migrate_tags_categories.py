#!/usr/bin/env python3
"""
Tag and Category Migration Script for markgo

This script automates the consolidation of tags and categories across all markdown articles
based on the optimization plan defined in tag-category-optimization-plan.md.

Usage:
    python migrate_tags_categories.py [options]

Options:
    --dry-run    : Show what changes would be made without actually modifying files
    --backup     : Create backup of original files before modification
    --articles   : Path to articles directory (default: ./articles)
    --verbose    : Enable verbose output

Author: Generated for markgo optimization
"""

import os
import re
import sys
import argparse
import shutil
from pathlib import Path
from typing import Dict, List, Set, Tuple
import yaml


class TagCategoryMigrator:
    def __init__(self, articles_dir: str, dry_run: bool = False, backup: bool = False, verbose: bool = False):
        self.articles_dir = Path(articles_dir)
        self.dry_run = dry_run
        self.backup = backup
        self.verbose = verbose

        # Define category migration mappings
        self.category_mappings = {
            # Old category → New category
            'technology': 'technology',
            'Technology': 'technology',
            'General': 'technology',
            'personal': 'personal',
            'Personal': 'personal',
            'personal-reflection': 'personal',
            'thoughts': 'personal',
            'musings': 'personal',
            'career': 'career',
            'Career': 'career',
            'leadership': 'career',
            'music': 'music',
            'cricket': 'cricket',
            'philosophy': 'philosophy',
            'reflection': 'philosophy',
            'tools': 'tools',
            'development-tools': 'tools',
            'devops': 'tools',
            'DevOps': 'tools',
            'infrastructure': 'tools',
            'Infrastructure': 'tools',
            'culture': 'culture',
            'travel': 'culture',
            'entrepreneurship': 'culture',
            'community': 'culture',
            'database': 'technology',
            'backend': 'technology',
            'backend-development': 'technology',
            'javascript': 'technology',
            'guides': 'technology',
            'tutorials': 'technology',
            'events': 'culture',
            'reviews': 'tools',
            'search': 'technology',
            'announcements': 'technology',
            'Thoughts': 'personal',
        }

        # Define tag migration mappings
        self.tag_mappings = {
            # Programming Languages & Frameworks
            'go': 'golang',
            'go-programming': 'golang',
            'node-js': 'nodejs',

            # Architecture & Systems
            'microservices': 'distributed-systems',
            'circuit-breaker': 'distributed-systems',
            'resilience-patterns': 'distributed-systems',
            'architecture': 'system-architecture',
            'system-design': 'system-architecture',
            'backend-development': 'backend',
            'scalability': 'performance',  # Context-dependent, may need manual review
            'disaster-recovery': 'high-availability',

            # Data & Storage
            'postgresql': 'database',
            'mysql': 'database',
            'streaming-replication': 'database',
            'distributed-caching': 'redis',
            'caching': 'redis',
            'statistics': 'data-analysis',
            'sports-analytics': 'data-analysis',

            # Infrastructure & Operations
            'devops': 'infrastructure',
            'production': 'infrastructure',
            'deployment': 'infrastructure',
            'monitoring': 'infrastructure',
            'load-balancing': 'nginx',
            'system-administration': 'linux',
            'xfs': 'linux',
            'bash': 'automation',
            'backup': 'automation',

            # Development Practices
            'development-standards': 'best-practices',
            'maintainability': 'best-practices',
            'code-style': 'best-practices',
            'version-control': 'git',
            'branching-strategy': 'git',
            'rebase': 'git',
            'code-reviews': 'tools-review',
            'bug-tracking': 'tools-review',
            'team-collaboration': 'tools-review',

            # Content Types
            'getting-started': 'tutorial',
            'guides': 'tutorial',
            'tech-tools': 'tools-review',
            'review': 'tools-review',
            'development-tools': 'tools-review',

            # Specialized Topics
            'search-engine': 'search',
            'information-retrieval': 'search',
            'elasticsearch': 'search',
            'dynamic-dns': 'networking',
            'wifi-configuration': 'networking',
            'raspberry-pi': 'embedded-systems',
            'home-server': 'embedded-systems',
            'array-sorting': 'algorithms',
            'data-manipulation': 'algorithms',
            'frontend-development': 'web-development',

            # Personal & Career
            'career': 'career-journey',
            'career-transition': 'career-journey',
            'early-career': 'career-journey',
            'career-advice': 'personal-growth',
            'life-lessons': 'personal-growth',
            'self-reflection': 'personal-growth',
            'authenticity': 'personal-growth',
            'mentorship': 'leadership',
            'teamwork': 'leadership',
            'engineering-management': 'leadership',
            'startup-life': 'entrepreneurship',
            'startup-ecosystem': 'entrepreneurship',
            'startup': 'entrepreneurship',
            'holidays': 'work-life-balance',
            'team-building': 'work-life-balance',
            'reflections': 'reflection',
            'personal-reflection': 'reflection',
            'introspection': 'reflection',
            'existentialism': 'philosophy',
            'meaning': 'philosophy',
            'consciousness': 'philosophy',
            'motivation': 'life-lessons',
            'inspiration': 'life-lessons',
            'career-development': 'impostor-syndrome',
            'system-design': 'interviews',
            'faang': 'interviews',
            'senior-engineer': 'interviews',

            # Cultural & Interest
            'music': 'music-analysis',
            'musical-analysis': 'music-analysis',
            'storytelling': 'music-analysis',  # Context-dependent
            'decibel': 'indian-rock',
            'naagin': 'indian-rock',
            'cultural-fusion': 'indian-rock',
            'rock-music': 'classic-rock',
            'bryan-adams': 'classic-rock',
            'bob-dylan': 'classic-rock',
            'live-concerts': 'music-events',
            'india-tour': 'music-events',
            'cricket': 'cricket-analysis',
            'cricket-history': 'cricket-analysis',
            'kathmandu': 'nepal',
            'nepali-poetry': 'nepal',
            'cultural-identity': 'nepal',
            'hyderabad': 'india',
            'delhi': 'india',
            'nizam-era': 'india',
            'heritage': 'travel',
            'photography': 'travel',
            'history': 'travel',
            'metaverse': 'future-tech',
            'artificial-intelligence': 'future-tech',
            'virtual-reality': 'future-tech',
            'gnome': 'open-source',
            'legacy': 'open-source',
            'english': 'language',
            'poetry': 'language',
            'pronunciation': 'language',
            'literature': 'language',
            'gource': 'visualization',
            'ffmpeg': 'visualization',
            'video-creation': 'visualization',

            # Specialized
            'diffusion': 'phabricator',
            'chaos': 'humor',
            'freedom': 'rant',
            'individuality': 'rant',
            'pink-floyd': 'rant',
            'spirituality': 'mythology',
            'ancient-wisdom': 'mythology',
            'blog': 'announcements',
            'migration': 'announcements',
            'fresh-start': 'announcements',
            'welcome': 'announcements',
            'web-design': 'news-aggregation',  # Context-dependent
            'zyoba-labs': 'community',
            'blogging': 'community',

            # Tags to remove (too specific or redundant)
            'user-experience': None,  # Remove - too broad
            'thamel': None,          # Remove - too specific
            'purple-haze': None,     # Remove - too specific
            'new-year': None,        # Remove - too specific
            'farewell': None,        # Remove - too specific
            'birthdays': None,       # Remove - too specific
            'song-of-the-day': None, # Remove - too specific
        }

        # Track statistics
        self.stats = {
            'files_processed': 0,
            'files_modified': 0,
            'tags_consolidated': 0,
            'categories_consolidated': 0,
            'tags_removed': 0,
        }

    def log(self, message: str, verbose_only: bool = False):
        """Log message with optional verbose filtering"""
        if not verbose_only or self.verbose:
            print(message)

    def parse_frontmatter(self, content: str) -> Tuple[Dict, str]:
        """Parse YAML frontmatter from markdown content"""
        if not content.startswith('---'):
            return {}, content

        parts = content.split('---', 2)
        if len(parts) < 3:
            return {}, content

        try:
            # Parse frontmatter but preserve original date strings
            frontmatter_text = parts[1]
            frontmatter = yaml.safe_load(frontmatter_text)
            body = parts[2]

            # Extract original date string if present
            if frontmatter and 'date' in frontmatter:
                for line in frontmatter_text.split('\n'):
                    line = line.strip()
                    if line.startswith('date:'):
                        # Extract the original date string
                        date_part = line.split('date:', 1)[1].strip()
                        frontmatter['date'] = date_part
                        break

            return frontmatter or {}, body
        except yaml.YAMLError as e:
            self.log(f"Error parsing YAML frontmatter: {e}")
            return {}, content

    def serialize_frontmatter(self, frontmatter: Dict, body: str) -> str:
        """Serialize frontmatter and body back to markdown"""
        if not frontmatter:
            return body

        # Simple approach: read original content and only replace tags/categories lines
        lines = ["---"]

        for key, value in frontmatter.items():
            if key in ['tags', 'categories'] and isinstance(value, list):
                # Format as square bracket CSV: [tag1, tag2, tag3]
                if value:
                    csv_value = ', '.join(value)
                    lines.append(f"{key}: [{csv_value}]")
                else:
                    lines.append(f"{key}: []")
            elif key == 'date':
                # Preserve date field exactly as string to avoid format changes
                lines.append(f"{key}: {value}")
            else:
                # Use standard YAML serialization for all other fields
                yaml_line = yaml.dump({key: value}, default_flow_style=False, sort_keys=False).strip()
                lines.append(yaml_line)

        lines.append("---")
        return '\n'.join(lines) + body

    def migrate_categories(self, categories: List[str]) -> List[str]:
        """Migrate and consolidate categories"""
        if not categories:
            return categories

        new_categories = []
        for category in categories:
            if category in self.category_mappings:
                new_cat = self.category_mappings[category]
                if new_cat not in new_categories:
                    new_categories.append(new_cat)
                    if new_cat != category:
                        self.stats['categories_consolidated'] += 1
            else:
                # Keep unmapped categories as-is
                if category not in new_categories:
                    new_categories.append(category)

        return new_categories

    def migrate_tags(self, tags: List[str]) -> List[str]:
        """Migrate and consolidate tags"""
        if not tags:
            return tags

        new_tags = []
        for tag in tags:
            if tag in self.tag_mappings:
                new_tag = self.tag_mappings[tag]
                if new_tag is None:
                    # Tag marked for removal
                    self.stats['tags_removed'] += 1
                    continue
                elif new_tag not in new_tags:
                    new_tags.append(new_tag)
                    if new_tag != tag:
                        self.stats['tags_consolidated'] += 1
            else:
                # Keep unmapped tags as-is
                if tag not in new_tags:
                    new_tags.append(tag)

        return new_tags

    def process_file(self, file_path: Path) -> bool:
        """Process a single markdown file"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()

            frontmatter, body = self.parse_frontmatter(content)
            if not frontmatter:
                self.log(f"No frontmatter found in {file_path}", verbose_only=True)
                return False

            original_categories = frontmatter.get('categories', [])
            original_tags = frontmatter.get('tags', [])

            # Migrate categories and tags
            new_categories = self.migrate_categories(original_categories)
            new_tags = self.migrate_tags(original_tags)

            # Check if changes were made
            categories_changed = original_categories != new_categories
            tags_changed = original_tags != new_tags

            if not categories_changed and not tags_changed:
                self.log(f"No changes needed for {file_path.name}", verbose_only=True)
                return False

            # Update frontmatter
            if categories_changed:
                frontmatter['categories'] = new_categories
                self.log(f"Categories: {original_categories} → {new_categories}")

            if tags_changed:
                frontmatter['tags'] = new_tags
                self.log(f"Tags: {original_tags} → {new_tags}")

            if self.dry_run:
                self.log(f"[DRY RUN] Would update {file_path.name}")
                return True

            # Create backup if requested
            if self.backup:
                backup_path = file_path.with_suffix(f"{file_path.suffix}.backup")
                shutil.copy2(file_path, backup_path)
                self.log(f"Created backup: {backup_path}", verbose_only=True)

            # Write updated content
            updated_content = self.serialize_frontmatter(frontmatter, body)
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(updated_content)

            self.log(f"Updated {file_path.name}")
            return True

        except Exception as e:
            self.log(f"Error processing {file_path}: {e}")
            return False

    def run(self):
        """Run the migration process"""
        if not self.articles_dir.exists():
            self.log(f"Articles directory not found: {self.articles_dir}")
            return False

        markdown_files = list(self.articles_dir.glob("*.markdown")) + list(self.articles_dir.glob("*.md"))

        if not markdown_files:
            self.log(f"No markdown files found in {self.articles_dir}")
            return False

        self.log(f"Found {len(markdown_files)} markdown files to process")
        if self.dry_run:
            self.log("Running in DRY RUN mode - no files will be modified")

        for file_path in markdown_files:
            self.stats['files_processed'] += 1
            if self.process_file(file_path):
                self.stats['files_modified'] += 1

        # Print summary
        self.log("\n" + "="*50)
        self.log("MIGRATION SUMMARY")
        self.log("="*50)
        self.log(f"Files processed: {self.stats['files_processed']}")
        self.log(f"Files modified: {self.stats['files_modified']}")
        self.log(f"Tags consolidated: {self.stats['tags_consolidated']}")
        self.log(f"Categories consolidated: {self.stats['categories_consolidated']}")
        self.log(f"Tags removed: {self.stats['tags_removed']}")

        if self.dry_run:
            self.log("\nThis was a DRY RUN. Re-run without --dry-run to apply changes.")

        return True


def main():
    parser = argparse.ArgumentParser(
        description="Migrate and consolidate tags and categories in blog articles",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Dry run to see what changes would be made
  python migrate_tags_categories.py --dry-run --verbose

  # Apply changes with backup
  python migrate_tags_categories.py --backup --verbose

  # Apply changes to specific directory
  python migrate_tags_categories.py --articles /path/to/articles
        """
    )

    parser.add_argument(
        '--articles',
        default='./articles',
        help='Path to articles directory (default: ./articles)'
    )

    parser.add_argument(
        '--dry-run',
        action='store_true',
        help='Show what changes would be made without modifying files'
    )

    parser.add_argument(
        '--backup',
        action='store_true',
        help='Create backup files before modification'
    )

    parser.add_argument(
        '--verbose',
        action='store_true',
        help='Enable verbose output'
    )

    args = parser.parse_args()

    migrator = TagCategoryMigrator(
        articles_dir=args.articles,
        dry_run=args.dry_run,
        backup=args.backup,
        verbose=args.verbose
    )

    if migrator.run():
        sys.exit(0)
    else:
        sys.exit(1)


if __name__ == "__main__":
    main()
