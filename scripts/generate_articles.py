#!/usr/bin/env python3
"""
Generate varied articles for MarkGo testing
Creates medium to large articles with diverse content and topics

Usage:
    python3 scripts/generate_articles.py [number_of_articles] [output_directory]
    
Examples:
    python3 scripts/generate_articles.py 100 temp_articles
    python3 scripts/generate_articles.py 500 test_content
"""

import os
import sys
import random
from datetime import datetime, timedelta
from typing import List, Dict, Any

# Article topics and categories
CATEGORIES = [
    "Technology", "Programming", "Web Development", "DevOps", "Cloud Computing",
    "Artificial Intelligence", "Machine Learning", "Data Science", "Cybersecurity",
    "Mobile Development", "Frontend", "Backend", "Database", "Architecture",
    "Tools", "Productivity", "Career", "Tutorial", "Review", "Opinion",
    "Industry News", "Open Source", "Performance", "Testing", "Debugging"
]

TECH_TOPICS = [
    "JavaScript", "Python", "Go", "Rust", "TypeScript", "React", "Vue.js", "Angular",
    "Node.js", "Docker", "Kubernetes", "AWS", "Azure", "GCP", "MongoDB", "PostgreSQL",
    "Redis", "Elasticsearch", "GraphQL", "REST API", "Microservices", "Serverless",
    "CI/CD", "Git", "Linux", "Bash", "Vim", "VS Code", "IntelliJ", "Terraform",
    "Ansible", "Jenkins", "GitHub Actions", "GitLab", "Nginx", "Apache", "MySQL",
    "SQLite", "Firebase", "Supabase", "Vercel", "Netlify", "Heroku", "DigitalOcean"
]

FRAMEWORKS = [
    "Spring Boot", "Django", "Flask", "FastAPI", "Express.js", "Gin", "Fiber",
    "Laravel", "Rails", "Phoenix", "Next.js", "Nuxt.js", "Svelte", "Remix",
    "Astro", "Gatsby", "Hugo", "Jekyll", "Eleventy", "Tailwind CSS", "Bootstrap",
    "Material UI", "Chakra UI", "Ant Design", "Styled Components", "Sass", "LESS"
]

TOPICS_POOL = [
    # Programming Languages
    "The Evolution of {lang} in Modern Development",
    "Why {lang} is Perfect for {type} Applications",
    "Advanced {lang} Techniques Every Developer Should Know",
    "Building Scalable Applications with {lang}",
    "Performance Optimization in {lang}",
    "From Beginner to Expert: {lang} Learning Path",
    "Common {lang} Mistakes and How to Avoid Them",
    "Testing Strategies for {lang} Applications",
    
    # Web Development
    "Modern {framework} Development Best Practices",
    "Building Production-Ready {framework} Applications",
    "State Management in {framework}: A Complete Guide",
    "Performance Optimization for {framework} Apps",
    "Security Best Practices in {framework}",
    "From Zero to Production: {framework} Deployment Guide",
    "Advanced {framework} Patterns and Techniques",
    "Comparing {framework} with Alternative Solutions",
    
    # DevOps & Infrastructure
    "Container Orchestration with {tool}",
    "Infrastructure as Code using {tool}",
    "CI/CD Pipeline Design with {tool}",
    "Monitoring and Observability in {environment}",
    "Security Hardening for {platform}",
    "Cost Optimization Strategies for {cloud}",
    "Disaster Recovery Planning with {platform}",
    "Automated Testing in {environment}",
    
    # Career & Industry
    "Career Growth in {field}",
    "Essential Skills for {role} Developers",
    "Remote Work Best Practices for {role}",
    "Interview Preparation for {role} Positions",
    "Building a Portfolio as a {role}",
    "Networking Strategies in the {field} Industry",
    "Salary Negotiation for {role} Professionals",
    "Transitioning from {old_role} to {new_role}",
    
    # Tutorials & How-To
    "Step-by-Step Guide to {task}",
    "Building {project} from Scratch",
    "Implementing {feature} in {tech}",
    "Debugging {issue} in {environment}",
    "Optimizing {metric} for {application}",
    "Migrating from {old_tech} to {new_tech}",
    "Setting up {environment} for Development",
    "Automating {process} with {tool}",
    
    # Reviews & Comparisons
    "Comprehensive Review of {tool}",
    "Comparing {tool1} vs {tool2} in 2025",
    "Why We Switched from {old_tool} to {new_tool}",
    "Honest Review: {product} After 6 Months",
    "The Ultimate {category} Tool Comparison",
    "Is {tool} Worth the Hype? A Developer's Perspective",
    "Benchmarking {tool1} vs {tool2} Performance",
    "Feature Comparison: {tool} Alternatives",
    
    # Industry & Trends
    "The Future of {technology} in 2026",
    "How {trend} is Changing Software Development",
    "Industry Report: State of {field} in 2025",
    "Emerging Trends in {domain}",
    "The Impact of {technology} on {industry}",
    "Predictions for {field} in the Next Decade",
    "Why {concept} is the Next Big Thing",
    "The Rise and Fall of {technology}",
]

# Content templates for different article types
TUTORIAL_SECTIONS = [
    "## Prerequisites\n\nBefore we begin, make sure you have:",
    "## Setting Up the Environment\n\nFirst, let's set up our development environment:",
    "## Step-by-Step Implementation\n\nNow let's implement the solution step by step:",
    "## Best Practices\n\nHere are some best practices to keep in mind:",
    "## Common Issues and Solutions\n\nYou might encounter these common issues:",
    "## Testing and Validation\n\nLet's test our implementation:",
    "## Performance Considerations\n\nFor optimal performance, consider:",
    "## Security Considerations\n\nFrom a security perspective:",
    "## Deployment and Production\n\nWhen deploying to production:",
    "## Conclusion\n\nIn this tutorial, we've covered:",
]

REVIEW_SECTIONS = [
    "## Introduction\n\nIn this comprehensive review, we'll examine:",
    "## Key Features\n\nThe standout features include:",
    "## Performance Analysis\n\nOur performance testing revealed:",
    "## Pros and Cons\n\n### Advantages:",
    "### Disadvantages:",
    "## Use Cases\n\nThis tool excels in scenarios such as:",
    "## Comparison with Alternatives\n\nCompared to similar tools:",
    "## Pricing and Value\n\nFrom a cost perspective:",
    "## Community and Support\n\nThe community ecosystem offers:",
    "## Final Verdict\n\nAfter extensive testing:",
]

OPINION_SECTIONS = [
    "## The Current State of Affairs\n\nLooking at the current landscape:",
    "## Why This Matters\n\nThis topic is crucial because:",
    "## Different Perspectives\n\nThere are several viewpoints to consider:",
    "## Personal Experience\n\nFrom my own experience:",
    "## Industry Implications\n\nThe broader implications include:",
    "## Future Outlook\n\nLooking ahead, we can expect:",
    "## Recommendations\n\nBased on this analysis, I recommend:",
    "## Call to Action\n\nWhat can we do about this?",
]

def generate_random_content(sections: List[str], min_paragraphs: int = 3, max_paragraphs: int = 8) -> str:
    """Generate random content using the provided sections"""
    content = ""
    
    for section in random.sample(sections, random.randint(min(4, len(sections)), len(sections))):
        content += section + "\n\n"
        
        # Add 2-5 paragraphs per section
        num_paragraphs = random.randint(min_paragraphs, max_paragraphs)
        for _ in range(num_paragraphs):
            content += generate_paragraph() + "\n\n"
    
    return content

def generate_paragraph() -> str:
    """Generate a realistic paragraph of technical content"""
    sentences = []
    num_sentences = random.randint(3, 7)
    
    for _ in range(num_sentences):
        sentences.append(generate_sentence())
    
    return " ".join(sentences)

def generate_sentence() -> str:
    """Generate a realistic technical sentence"""
    sentence_templates = [
        "This approach provides {benefit} while maintaining {quality}.",
        "When implementing {feature}, it's important to consider {consideration}.",
        "The {technology} ecosystem offers {advantage} for {use_case}.",
        "Many developers overlook {aspect} when working with {tool}.",
        "Performance benchmarks show {metric} improvement over {comparison}.",
        "The key to successful {process} lies in {factor}.",
        "Modern {field} practices emphasize {principle} and {value}.",
        "Integration with {service} enables {capability} across {scope}.",
        "Security researchers recommend {practice} to prevent {threat}.",
        "The latest version introduces {feature} for better {outcome}.",
        "Community feedback indicates strong preference for {approach}.",
        "Documentation clearly outlines the steps for {process}.",
        "Error handling becomes crucial when dealing with {scenario}.",
        "The configuration file should specify {parameter} for optimal {result}.",
        "Testing frameworks provide {functionality} to ensure {quality}."
    ]
    
    benefits = ["better performance", "improved scalability", "enhanced security", "greater flexibility", 
               "reduced complexity", "faster development", "better maintainability", "improved user experience"]
    qualities = ["code quality", "system stability", "data integrity", "user privacy", "application security"]
    technologies = ["React", "Node.js", "Docker", "Kubernetes", "PostgreSQL", "Redis", "GraphQL", "TypeScript"]
    features = ["authentication", "caching", "routing", "state management", "data validation", "error tracking"]
    considerations = ["performance implications", "security vulnerabilities", "scalability requirements", "browser compatibility"]
    
    template = random.choice(sentence_templates)
    
    # Fill in template placeholders
    replacements = {
        "{benefit}": random.choice(benefits),
        "{quality}": random.choice(qualities),
        "{technology}": random.choice(technologies),
        "{feature}": random.choice(features),
        "{consideration}": random.choice(considerations),
        "{tool}": random.choice(TECH_TOPICS),
        "{advantage}": random.choice(benefits),
        "{use_case}": random.choice(["web applications", "mobile apps", "APIs", "microservices", "data processing"]),
        "{aspect}": random.choice(["error handling", "performance optimization", "security", "testing", "documentation"]),
        "{metric}": f"{random.randint(15, 85)}%",
        "{comparison}": "previous versions",
        "{process}": random.choice(["deployment", "testing", "development", "debugging", "optimization"]),
        "{factor}": random.choice(["proper planning", "team collaboration", "clear documentation", "automated testing"]),
        "{field}": random.choice(["software development", "web development", "DevOps", "data engineering"]),
        "{principle}": random.choice(["DRY principles", "SOLID principles", "clean code", "test-driven development"]),
        "{value}": random.choice(["maintainability", "readability", "performance", "security"]),
        "{service}": random.choice(["AWS", "Azure", "Google Cloud", "GitHub", "Docker Hub"]),
        "{capability}": random.choice(["seamless scaling", "automated deployment", "real-time monitoring", "data synchronization"]),
        "{scope}": random.choice(["multiple environments", "different platforms", "various devices", "global regions"]),
        "{practice}": random.choice(["input validation", "secure authentication", "encrypted communication", "regular updates"]),
        "{threat}": random.choice(["XSS attacks", "SQL injection", "data breaches", "unauthorized access"]),
        "{outcome}": random.choice(["performance", "reliability", "security", "usability"]),
        "{approach}": random.choice(["declarative syntax", "functional programming", "microservices architecture", "containerization"]),
        "{functionality}": random.choice(["mocking capabilities", "assertion libraries", "coverage reporting", "parallel execution"]),
        "{scenario}": random.choice(["network failures", "high traffic", "data corruption", "service outages"]),
        "{parameter}": random.choice(["timeout values", "cache expiration", "connection pools", "retry policies"]),
        "{result}": random.choice(["performance", "reliability", "efficiency", "throughput"])
    }
    
    for placeholder, value in replacements.items():
        template = template.replace(placeholder, value)
    
    return template

def generate_code_snippet(language: str) -> str:
    """Generate a realistic code snippet"""
    if language.lower() in ["javascript", "js", "typescript", "ts"]:
        return '''```javascript
// Example implementation
const fetchUserData = async (userId) => {
  try {
    const response = await fetch(`/api/users/${userId}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const userData = await response.json();
    return userData;
  } catch (error) {
    console.error('Error fetching user data:', error);
    throw error;
  }
};

// Usage
fetchUserData('123')
  .then(user => console.log('User data:', user))
  .catch(error => console.error('Failed to fetch user:', error));
```'''
    
    elif language.lower() in ["python", "py"]:
        return '''```python
import asyncio
import aiohttp
from typing import Dict, Any

async def fetch_user_data(user_id: str) -> Dict[str, Any]:
    """Fetch user data from API with proper error handling."""
    async with aiohttp.ClientSession() as session:
        try:
            async with session.get(f"/api/users/{user_id}") as response:
                response.raise_for_status()
                return await response.json()
        except aiohttp.ClientError as e:
            print(f"Error fetching user data: {e}")
            raise

# Usage
async def main():
    try:
        user_data = await fetch_user_data("123")
        print(f"User data: {user_data}")
    except Exception as e:
        print(f"Failed to fetch user: {e}")

if __name__ == "__main__":
    asyncio.run(main())
```'''
    
    elif language.lower() in ["go", "golang"]:
        return '''```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func fetchUserData(userID string) (*User, error) {
    client := &http.Client{
        Timeout: 10 * time.Second,
    }
    
    resp, err := client.Get(fmt.Sprintf("/api/users/%s", userID))
    if err != nil {
        return nil, fmt.Errorf("failed to fetch user: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
    }
    
    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &user, nil
}

func main() {
    user, err := fetchUserData("123")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("User: %+v\n", user)
}
```'''
    
    else:
        return '''```bash
#!/bin/bash

# Example deployment script
set -euo pipefail

DEPLOY_ENV=${1:-staging}
APP_NAME="my-application"

echo "Deploying $APP_NAME to $DEPLOY_ENV environment..."

# Build the application
echo "Building application..."
docker build -t $APP_NAME:latest .

# Tag for deployment
docker tag $APP_NAME:latest $APP_NAME:$DEPLOY_ENV

# Deploy
echo "Deploying to $DEPLOY_ENV..."
kubectl set image deployment/$APP_NAME container=$APP_NAME:$DEPLOY_ENV

# Wait for rollout
kubectl rollout status deployment/$APP_NAME

echo "Deployment completed successfully!"
```'''

def generate_article_content(title: str, category: str, article_type: str) -> str:
    """Generate comprehensive article content based on type"""
    content = ""
    
    # Introduction
    content += f"In this comprehensive guide, we'll explore {title.lower()}. "
    content += f"This {article_type} covers everything from basic concepts to advanced techniques, "
    content += f"providing practical insights for developers working in {category.lower()}.\n\n"
    
    # Add type-specific content
    if article_type == "tutorial":
        content += generate_random_content(TUTORIAL_SECTIONS, 4, 8)
        # Add code snippets
        languages = ["javascript", "python", "go", "bash"]
        for _ in range(random.randint(2, 4)):
            content += f"\n{generate_code_snippet(random.choice(languages))}\n\n"
            content += generate_paragraph() + "\n\n"
    
    elif article_type == "review":
        content += generate_random_content(REVIEW_SECTIONS, 3, 6)
    
    elif article_type == "opinion":
        content += generate_random_content(OPINION_SECTIONS, 4, 7)
    
    else:  # general article
        all_sections = TUTORIAL_SECTIONS + REVIEW_SECTIONS + OPINION_SECTIONS
        content += generate_random_content(random.sample(all_sections, random.randint(5, 8)), 3, 6)
        
        # Add occasional code snippet
        if random.random() < 0.6:  # 60% chance
            languages = ["javascript", "python", "go", "bash"]
            content += f"\n{generate_code_snippet(random.choice(languages))}\n\n"
    
    # Add conclusion
    content += "## Conclusion\n\n"
    content += generate_paragraph() + "\n\n"
    content += generate_paragraph() + "\n\n"
    
    # Add call to action
    if random.random() < 0.3:  # 30% chance
        content += "## What's Next?\n\n"
        content += generate_paragraph() + "\n\n"
    
    return content

def generate_frontmatter(title: str, category: str, date: datetime, tags: List[str], article_type: str) -> str:
    """Generate YAML frontmatter for article"""
    
    # Generate description
    descriptions = [
        f"A comprehensive guide to {title.lower()} covering best practices and real-world examples.",
        f"Learn about {title.lower()} with practical examples and expert insights.",
        f"Deep dive into {title.lower()} - from basics to advanced techniques.",
        f"Everything you need to know about {title.lower()} in modern development.",
        f"Practical guide to implementing {title.lower()} in your projects.",
        f"Master {title.lower()} with this detailed tutorial and examples.",
        f"Complete overview of {title.lower()} with hands-on examples.",
        f"Expert insights on {title.lower()} for modern developers."
    ]
    
    authors = [
        "Alex Chen", "Sarah Johnson", "Michael Rodriguez", "Emily Zhang", "David Kim",
        "Jessica Wong", "Ryan O'Connor", "Lisa Thompson", "Ahmed Hassan", "Maria Garcia",
        "James Wilson", "Priya Patel", "Tom Anderson", "Rachel Green", "Kevin Liu",
        "Anna Kowalski", "Carlos Mendez", "Sophie Martin", "Jake Peterson", "Nina Popov"
    ]
    
    frontmatter = f"""---
title: "{title}"
description: "{random.choice(descriptions)}"
date: {date.strftime('%Y-%m-%dT%H:%M:%SZ')}
tags: {tags}
categories: ["{category}"]
featured: {str(random.random() < 0.15).lower()}
draft: {str(random.random() < 0.1).lower()}
author: "{random.choice(authors)}"
reading_time: {random.randint(5, 25)} min
seo_title: "{title} - Complete Guide"
seo_description: "Learn {title.lower()} with practical examples, best practices, and expert insights. Comprehensive tutorial for developers."
---

"""
    return frontmatter

def generate_filename(title: str, date: datetime) -> str:
    """Generate filename from title and date"""
    # Clean title for filename
    clean_title = title.lower()
    clean_title = ''.join(c if c.isalnum() or c in ' -' else '' for c in clean_title)
    clean_title = '-'.join(clean_title.split())
    clean_title = clean_title[:80]  # Limit length
    
    return f"{date.strftime('%Y-%m-%d')}-{clean_title}.md"

def generate_articles(num_articles: int = 500, output_dir: str = "temp_articles") -> None:
    """Generate the specified number of articles"""
    
    # Ensure output directory exists
    os.makedirs(output_dir, exist_ok=True)
    
    # Generate date range (last 3 years)
    end_date = datetime.now()
    start_date = end_date - timedelta(days=1095)  # ~3 years
    
    article_types = ["tutorial", "review", "opinion", "guide", "analysis", "comparison"]
    
    print(f"Generating {num_articles} articles in {output_dir}/...")
    
    for i in range(num_articles):
        # Generate random date
        random_days = random.randint(0, 1095)
        article_date = start_date + timedelta(days=random_days)
        
        # Select random elements
        category = random.choice(CATEGORIES)
        article_type = random.choice(article_types)
        
        # Generate title
        title_template = random.choice(TOPICS_POOL)
        
        # Create a comprehensive replacement dictionary
        replacements = {
            "lang": random.choice(["JavaScript", "Python", "Go", "TypeScript", "Rust"]),
            "type": random.choice(["web", "mobile", "enterprise", "cloud-native"]),
            "framework": random.choice(FRAMEWORKS),
            "tool": random.choice(TECH_TOPICS),
            "tool1": random.choice(TECH_TOPICS),
            "tool2": random.choice(TECH_TOPICS),
            "field": random.choice(["Software Development", "DevOps", "Data Science"]),
            "role": random.choice(["Frontend", "Backend", "Full-Stack", "DevOps"]),
            "old_role": random.choice(["Junior", "Mid-Level"]),
            "new_role": random.choice(["Senior", "Lead", "Principal"]),
            "task": random.choice(["building a REST API", "setting up CI/CD", "implementing authentication"]),
            "project": random.choice(["a blog engine", "a task manager", "an e-commerce site"]),
            "tech": random.choice(TECH_TOPICS),
            "environment": random.choice(["production", "development", "staging"]),
            "issue": random.choice(["memory leaks", "performance issues", "connection errors"]),
            "application": random.choice(["web applications", "mobile apps", "microservices"]),
            "metric": random.choice(["response time", "throughput", "memory usage"]),
            "old_tech": random.choice(TECH_TOPICS),
            "new_tech": random.choice(TECH_TOPICS),
            "process": random.choice(["deployment", "testing", "monitoring"]),
            "technology": random.choice(TECH_TOPICS + ["AI", "Machine Learning", "Blockchain"]),
            "trend": random.choice(["AI integration", "edge computing", "serverless architecture"]),
            "domain": random.choice(["web development", "mobile development", "cloud computing"]),
            "industry": random.choice(["fintech", "healthcare", "e-commerce"]),
            "concept": random.choice(["edge computing", "quantum computing", "Web3"]),
            "feature": random.choice(["authentication", "caching", "routing", "state management"]),
            "consideration": random.choice(["performance", "security", "scalability", "maintainability"]),
            "platform": random.choice(["AWS", "Azure", "Google Cloud", "Kubernetes"]),
            "cloud": random.choice(["AWS", "Azure", "Google Cloud"]),
            "category": random.choice(["development tools", "frameworks", "databases", "cloud services"]),
            "product": random.choice(TECH_TOPICS),
            "old_tool": random.choice(TECH_TOPICS),
            "new_tool": random.choice(TECH_TOPICS)
        }
        
        # Apply all replacements to handle any combination
        title = title_template
        for placeholder, value in replacements.items():
            title = title.replace(f"{{{placeholder}}}", value)
        
        # Generate tags
        base_tags = [category.lower().replace(" ", "-")]
        tech_tags = random.sample(TECH_TOPICS, random.randint(2, 5))
        base_tags.extend([tag.lower().replace(" ", "-") for tag in tech_tags])
        
        # Additional contextual tags
        if article_type == "tutorial":
            base_tags.extend(["tutorial", "step-by-step", "guide"])
        elif article_type == "review":
            base_tags.extend(["review", "analysis", "comparison"])
        elif article_type == "opinion":
            base_tags.extend(["opinion", "thoughts", "perspective"])
        
        tags = list(set(base_tags[:8]))  # Limit to 8 unique tags
        
        # Generate content
        frontmatter = generate_frontmatter(title, category, article_date, tags, article_type)
        content = generate_article_content(title, category, article_type)
        
        # Create filename
        filename = generate_filename(title, article_date)
        filepath = os.path.join(output_dir, filename)
        
        # Write article
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(frontmatter)
            f.write(content)
        
        if (i + 1) % 50 == 0:
            print(f"Generated {i + 1}/{num_articles} articles...")
    
    print(f"\nSuccessfully generated {num_articles} articles in {output_dir}/")
    print(f"Articles range from {start_date.strftime('%Y-%m-%d')} to {end_date.strftime('%Y-%m-%d')}")

def main():
    """Main function to handle command line arguments"""
    # Parse command line arguments
    num_articles = 500
    output_dir = "temp_articles"
    
    if len(sys.argv) > 1:
        try:
            num_articles = int(sys.argv[1])
        except ValueError:
            print(f"Error: '{sys.argv[1]}' is not a valid number")
            sys.exit(1)
    
    if len(sys.argv) > 2:
        output_dir = sys.argv[2]
    
    # Validate arguments
    if num_articles <= 0:
        print("Error: Number of articles must be positive")
        sys.exit(1)
    
    if num_articles > 10000:
        print("Warning: Generating more than 10,000 articles may take a long time")
        confirm = input("Continue? (y/N): ")
        if confirm.lower() != 'y':
            print("Cancelled")
            sys.exit(0)
    
    # Generate articles
    generate_articles(num_articles, output_dir)

if __name__ == "__main__":
    main()