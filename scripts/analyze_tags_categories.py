#!/usr/bin/env python3
"""
Tag and Category Analysis Script for markgo

This script analyzes the current state of tags and categories across all markdown articles
to provide insights for optimization and validate migration results.

Usage:
    python analyze_tags_categories.py [options]

Options:
    --articles   : Path to articles directory (default: ./articles)
    --output     : Output format: text, json, csv (default: text)
    --sort       : Sort by: frequency, alphabetical (default: frequency)
    --min-count  : Minimum usage count to include in report (default: 1)

Author: Generated for markgo optimization
"""

import os
import re
import sys
import argparse
import json
import csv
from pathlib import Path
from typing import Dict, List, Set, Tuple, Counter
from collections import Counter, defaultdict
import yaml


class TagCategoryAnalyzer:
    def __init__(self, articles_dir: str):
        self.articles_dir = Path(articles_dir)
        self.tags_counter = Counter()
        self.categories_counter = Counter()
        self.articles_data = []
        self.tag_cooccurrence = defaultdict(Counter)
        self.category_articles = defaultdict(list)
        self.tag_articles = defaultdict(list)

    def parse_frontmatter(self, content: str) -> Tuple[Dict, str]:
        """Parse YAML frontmatter from markdown content"""
        if not content.startswith('---'):
            return {}, content

        parts = content.split('---', 2)
        if len(parts) < 3:
            return {}, content

        try:
            frontmatter = yaml.safe_load(parts[1])
            body = parts[2]
            return frontmatter or {}, body
        except yaml.YAMLError as e:
            print(f"Error parsing YAML frontmatter: {e}")
            return {}, content

    def analyze_file(self, file_path: Path) -> Dict:
        """Analyze a single markdown file"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()

            frontmatter, body = self.parse_frontmatter(content)
            if not frontmatter:
                return {}

            tags = frontmatter.get('tags', [])
            categories = frontmatter.get('categories', [])
            title = frontmatter.get('title', file_path.stem)
            date = frontmatter.get('date', '')
            featured = frontmatter.get('featured', False)

            # Count tags and categories
            for tag in tags:
                self.tags_counter[tag] += 1
                self.tag_articles[tag].append(title)

            for category in categories:
                self.categories_counter[category] += 1
                self.category_articles[category].append(title)

            # Track tag co-occurrence
            for i, tag1 in enumerate(tags):
                for tag2 in tags[i+1:]:
                    self.tag_cooccurrence[tag1][tag2] += 1
                    self.tag_cooccurrence[tag2][tag1] += 1

            article_data = {
                'file': file_path.name,
                'title': title,
                'date': str(date),
                'featured': featured,
                'tags': tags,
                'categories': categories,
                'tag_count': len(tags),
                'category_count': len(categories)
            }

            self.articles_data.append(article_data)
            return article_data

        except Exception as e:
            print(f"Error analyzing {file_path}: {e}")
            return {}

    def run_analysis(self):
        """Run the complete analysis"""
        if not self.articles_dir.exists():
            print(f"Articles directory not found: {self.articles_dir}")
            return False

        markdown_files = list(self.articles_dir.glob("*.markdown")) + list(self.articles_dir.glob("*.md"))

        if not markdown_files:
            print(f"No markdown files found in {self.articles_dir}")
            return False

        print(f"Analyzing {len(markdown_files)} markdown files...")

        for file_path in markdown_files:
            self.analyze_file(file_path)

        return True

    def get_summary_stats(self) -> Dict:
        """Generate summary statistics"""
        total_articles = len(self.articles_data)
        total_tags = len(self.tags_counter)
        total_categories = len(self.categories_counter)

        avg_tags_per_article = sum(len(article['tags']) for article in self.articles_data) / total_articles if total_articles > 0 else 0
        avg_categories_per_article = sum(len(article['categories']) for article in self.articles_data) / total_articles if total_articles > 0 else 0

        # Find articles with most/least tags
        articles_by_tag_count = sorted(self.articles_data, key=lambda x: x['tag_count'], reverse=True)
        most_tagged = articles_by_tag_count[0] if articles_by_tag_count else None
        least_tagged = articles_by_tag_count[-1] if articles_by_tag_count else None

        # Find singleton tags (used only once)
        singleton_tags = [tag for tag, count in self.tags_counter.items() if count == 1]
        singleton_categories = [cat for cat, count in self.categories_counter.items() if count == 1]

        return {
            'total_articles': total_articles,
            'total_tags': total_tags,
            'total_categories': total_categories,
            'avg_tags_per_article': round(avg_tags_per_article, 2),
            'avg_categories_per_article': round(avg_categories_per_article, 2),
            'singleton_tags': len(singleton_tags),
            'singleton_categories': len(singleton_categories),
            'most_tagged_article': most_tagged,
            'least_tagged_article': least_tagged,
            'singleton_tag_list': singleton_tags,
            'singleton_category_list': singleton_categories
        }

    def find_similar_tags(self, threshold: float = 0.7) -> List[Tuple[str, str, float]]:
        """Find potentially similar tags based on string similarity"""
        from difflib import SequenceMatcher

        tags = list(self.tags_counter.keys())
        similar_pairs = []

        for i, tag1 in enumerate(tags):
            for tag2 in tags[i+1:]:
                similarity = SequenceMatcher(None, tag1.lower(), tag2.lower()).ratio()
                if similarity >= threshold:
                    similar_pairs.append((tag1, tag2, similarity))

        return sorted(similar_pairs, key=lambda x: x[2], reverse=True)

    def find_redundant_tags(self) -> Dict[str, List[str]]:
        """Find tags that might be redundant based on co-occurrence patterns"""
        redundant_candidates = {}

        for tag1, cooccurrences in self.tag_cooccurrence.items():
            if not cooccurrences:
                continue

            # Find tags that always appear together with tag1
            tag1_count = self.tags_counter[tag1]
            for tag2, cooccur_count in cooccurrences.items():
                # If tag2 appears in 80%+ of tag1's articles
                if cooccur_count >= tag1_count * 0.8:
                    if tag1 not in redundant_candidates:
                        redundant_candidates[tag1] = []
                    redundant_candidates[tag1].append(f"{tag2} (appears together {cooccur_count}/{tag1_count} times)")

        return redundant_candidates

    def generate_report(self, output_format: str = 'text', sort_by: str = 'frequency', min_count: int = 1):
        """Generate analysis report"""
        stats = self.get_summary_stats()

        if output_format == 'json':
            return self._generate_json_report(stats, sort_by, min_count)
        elif output_format == 'csv':
            return self._generate_csv_report(stats, sort_by, min_count)
        else:
            return self._generate_text_report(stats, sort_by, min_count)

    def _generate_text_report(self, stats: Dict, sort_by: str, min_count: int) -> str:
        """Generate text format report"""
        report = []
        report.append("=" * 60)
        report.append("TAG AND CATEGORY ANALYSIS REPORT")
        report.append("=" * 60)

        # Summary Statistics
        report.append("\nSUMMARY STATISTICS")
        report.append("-" * 30)
        report.append(f"Total Articles: {stats['total_articles']}")
        report.append(f"Total Unique Tags: {stats['total_tags']}")
        report.append(f"Total Unique Categories: {stats['total_categories']}")
        report.append(f"Average Tags per Article: {stats['avg_tags_per_article']}")
        report.append(f"Average Categories per Article: {stats['avg_categories_per_article']}")
        report.append(f"Singleton Tags (used once): {stats['singleton_tags']}")
        report.append(f"Singleton Categories (used once): {stats['singleton_categories']}")

        # Most/Least tagged articles
        if stats['most_tagged_article']:
            report.append(f"\nMost Tagged Article: '{stats['most_tagged_article']['title']}' ({stats['most_tagged_article']['tag_count']} tags)")
        if stats['least_tagged_article']:
            report.append(f"Least Tagged Article: '{stats['least_tagged_article']['title']}' ({stats['least_tagged_article']['tag_count']} tags)")

        # Tag frequency analysis
        report.append(f"\nTAG FREQUENCY ANALYSIS (min count: {min_count})")
        report.append("-" * 50)

        if sort_by == 'frequency':
            sorted_tags = self.tags_counter.most_common()
        else:
            sorted_tags = sorted(self.tags_counter.items())

        for tag, count in sorted_tags:
            if count >= min_count:
                report.append(f"{tag:30} : {count:3d} articles")

        # Category frequency analysis
        report.append(f"\nCATEGORY FREQUENCY ANALYSIS (min count: {min_count})")
        report.append("-" * 50)

        if sort_by == 'frequency':
            sorted_categories = self.categories_counter.most_common()
        else:
            sorted_categories = sorted(self.categories_counter.items())

        for category, count in sorted_categories:
            if count >= min_count:
                report.append(f"{category:30} : {count:3d} articles")

        # Singleton tags
        if stats['singleton_tag_list']:
            report.append(f"\nSINGLETON TAGS ({len(stats['singleton_tag_list'])} total)")
            report.append("-" * 30)
            for i, tag in enumerate(sorted(stats['singleton_tag_list'])):
                if i % 3 == 0:
                    report.append("")
                report[-1] = report[-1] + f"{tag:25}"

        # Singleton categories
        if stats['singleton_category_list']:
            report.append(f"\n\nSINGLETON CATEGORIES ({len(stats['singleton_category_list'])} total)")
            report.append("-" * 30)
            for category in sorted(stats['singleton_category_list']):
                report.append(f"  {category}")

        # Similar tags
        similar_tags = self.find_similar_tags()
        if similar_tags:
            report.append(f"\nPOTENTIALLY SIMILAR TAGS")
            report.append("-" * 30)
            for tag1, tag2, similarity in similar_tags[:10]:  # Top 10
                report.append(f"{tag1:20} <-> {tag2:20} (similarity: {similarity:.2f})")

        # Redundant tags
        redundant_tags = self.find_redundant_tags()
        if redundant_tags:
            report.append(f"\nPOTENTIAL TAG REDUNDANCIES")
            report.append("-" * 30)
            for tag, cooccurring_tags in list(redundant_tags.items())[:10]:  # Top 10
                report.append(f"{tag}:")
                for cooccur in cooccurring_tags:
                    report.append(f"  -> {cooccur}")

        # Top tag combinations
        report.append(f"\nTOP TAG CO-OCCURRENCES")
        report.append("-" * 30)
        all_cooccurrences = []
        for tag1, cooccurrences in self.tag_cooccurrence.items():
            for tag2, count in cooccurrences.items():
                if tag1 < tag2:  # Avoid duplicates
                    all_cooccurrences.append((tag1, tag2, count))

        top_cooccurrences = sorted(all_cooccurrences, key=lambda x: x[2], reverse=True)[:10]
        for tag1, tag2, count in top_cooccurrences:
            report.append(f"{tag1:20} + {tag2:20} : {count} times")

        return "\n".join(report)

    def _generate_json_report(self, stats: Dict, sort_by: str, min_count: int) -> str:
        """Generate JSON format report"""
        data = {
            'summary': stats,
            'tags': dict(self.tags_counter.most_common() if sort_by == 'frequency' else sorted(self.tags_counter.items())),
            'categories': dict(self.categories_counter.most_common() if sort_by == 'frequency' else sorted(self.categories_counter.items())),
            'similar_tags': self.find_similar_tags(),
            'redundant_tags': self.find_redundant_tags(),
            'articles': self.articles_data
        }
        return json.dumps(data, indent=2)

    def _generate_csv_report(self, stats: Dict, sort_by: str, min_count: int) -> str:
        """Generate CSV format report"""
        # For simplicity, just return tag/category frequencies
        import io
        output = io.StringIO()

        # Tags CSV
        writer = csv.writer(output)
        writer.writerow(['type', 'name', 'count'])

        if sort_by == 'frequency':
            sorted_tags = self.tags_counter.most_common()
            sorted_categories = self.categories_counter.most_common()
        else:
            sorted_tags = sorted(self.tags_counter.items())
            sorted_categories = sorted(self.categories_counter.items())

        for tag, count in sorted_tags:
            if count >= min_count:
                writer.writerow(['tag', tag, count])

        for category, count in sorted_categories:
            if count >= min_count:
                writer.writerow(['category', category, count])

        return output.getvalue()


def main():
    parser = argparse.ArgumentParser(
        description="Analyze tags and categories in blog articles",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )

    parser.add_argument(
        '--articles',
        default='./articles',
        help='Path to articles directory (default: ./articles)'
    )

    parser.add_argument(
        '--output',
        choices=['text', 'json', 'csv'],
        default='text',
        help='Output format (default: text)'
    )

    parser.add_argument(
        '--sort',
        choices=['frequency', 'alphabetical'],
        default='frequency',
        help='Sort order (default: frequency)'
    )

    parser.add_argument(
        '--min-count',
        type=int,
        default=1,
        help='Minimum usage count to include in report (default: 1)'
    )

    parser.add_argument(
        '--output-file',
        help='Save report to file instead of printing to stdout'
    )

    args = parser.parse_args()

    analyzer = TagCategoryAnalyzer(args.articles)

    if not analyzer.run_analysis():
        sys.exit(1)

    report = analyzer.generate_report(
        output_format=args.output,
        sort_by=args.sort,
        min_count=args.min_count
    )

    if args.output_file:
        with open(args.output_file, 'w', encoding='utf-8') as f:
            f.write(report)
        print(f"Report saved to {args.output_file}")
    else:
        print(report)


if __name__ == "__main__":
    main()
