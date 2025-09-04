package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	defaultTitle       = "Untitled Article"
	defaultDescription = ""
	defaultTags        = "general"
	defaultCategory    = "uncategorized"
	defaultDraft       = true
	defaultFeatured    = false
	articlesDir        = "articles"
)

var (
	title       = flag.String("title", defaultTitle, "Article title")
	description = flag.String("description", defaultDescription, "Article description")
	tags        = flag.String("tags", defaultTags, "Comma-separated tags")
	category    = flag.String("category", defaultCategory, "Article category")
	author      = flag.String("author", "", "Author name (default: current OS username)")
	draft       = flag.Bool("draft", defaultDraft, "Mark article as draft")
	featured    = flag.Bool("featured", defaultFeatured, "Mark article as featured")
	help        = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Check if we should run interactive mode
	if shouldRunInteractive() {
		runInteractiveMode()
	}

	// Set default author if not provided
	if *author == "" {
		*author = getDefaultAuthor()
	}

	// Generate filename from title
	filename := slugify(*title) + ".markdown"
	filepath := filepath.Join(articlesDir, filename)

	// Check if file already exists
	if _, err := os.Stat(filepath); err == nil {
		fmt.Printf("Error: File %s already exists\n", filepath)
		os.Exit(1)
	}

	// Ensure articles directory exists
	if err := os.MkdirAll(articlesDir, 0755); err != nil {
		fmt.Printf("Error creating articles directory: %v\n", err)
		os.Exit(1)
	}

	// Generate and write article content
	content := generateArticle(*title, *description, *tags, *category, *author, *draft, *featured)
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	// Show success message
	fmt.Printf("✅ Article created: %s\n", filepath)
	fmt.Printf("📝 Title: %s\n", *title)
	fmt.Printf("👤 Author: %s\n", *author)
	fmt.Printf("🏷️  Tags: %s\n", *tags)
	fmt.Printf("📁 Category: %s\n", *category)
	fmt.Printf("📄 Draft: %v\n", *draft)
	fmt.Printf("⭐ Featured: %v\n", *featured)
}

func shouldRunInteractive() bool {
	// Run interactive if no flags were provided
	flagsProvided := false
	flag.Visit(func(f *flag.Flag) {
		flagsProvided = true
	})
	return !flagsProvided
}

func runInteractiveMode() {
	fmt.Println("🚀 Interactive Article Creator")
	fmt.Println("Press Enter to use defaults shown in [brackets]")
	fmt.Println()

	defaultAuthor := getDefaultAuthor()

	// Check if input is piped
	stat, _ := os.Stdin.Stat()
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	var inputs []string
	if isPiped {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputs = append(inputs, scanner.Text())
		}
	}

	// Get all inputs
	*title = getInputWithPipe("Title", defaultTitle, inputs, 0, isPiped)
	*description = getInputWithPipe("Description", defaultDescription, inputs, 1, isPiped)
	*tags = getInputWithPipe("Tags (comma-separated)", defaultTags, inputs, 2, isPiped)
	*category = getInputWithPipe("Category", defaultCategory, inputs, 3, isPiped)
	*author = getInputWithPipe("Author", defaultAuthor, inputs, 4, isPiped)
	*draft = getBoolInputWithPipe("Draft", defaultDraft, inputs, 5, isPiped)
	*featured = getBoolInputWithPipe("Featured", defaultFeatured, inputs, 6, isPiped)

	fmt.Println()
}

func getInput(prompt, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)

	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}

	return input
}

func getBoolInput(prompt string, defaultValue bool) bool {
	reader := bufio.NewReader(os.Stdin)
	defaultStr := "false"
	if defaultValue {
		defaultStr = "true"
	}

	for {
		fmt.Printf("%s (true/false) [%s]: ", prompt, defaultStr)

		input, err := reader.ReadString('\n')
		if err != nil {
			return defaultValue
		}

		input = strings.TrimSpace(strings.ToLower(input))
		if input == "" {
			return defaultValue
		}

		switch input {
		case "true", "t", "yes", "y", "1":
			return true
		case "false", "f", "no", "n", "0":
			return false
		default:
			fmt.Println("Please enter 'true' or 'false' (or press Enter for default)")
		}
	}
}

func getInputWithPipe(prompt, defaultValue string, inputs []string, index int, isPiped bool) string {
	if isPiped && index < len(inputs) {
		input := strings.TrimSpace(inputs[index])
		if input != "" {
			return input
		}
		return defaultValue
	}
	return getInput(prompt, defaultValue)
}

func getBoolInputWithPipe(prompt string, defaultValue bool, inputs []string, index int, isPiped bool) bool {
	if isPiped && index < len(inputs) {
		input := strings.TrimSpace(strings.ToLower(inputs[index]))
		if input != "" {
			switch input {
			case "true", "t", "yes", "y", "1":
				return true
			case "false", "f", "no", "n", "0":
				return false
			}
		}
		return defaultValue
	}
	return getBoolInput(prompt, defaultValue)
}

func getDefaultAuthor() string {
	if currentUser, err := user.Current(); err == nil {
		return currentUser.Username
	}
	return "Unknown Author"
}

func showHelp() {
	fmt.Println("new-article - Generate markdown blog articles with YAML frontmatter")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  new-article [OPTIONS]")
	fmt.Println("  new-article                    # Interactive mode")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Printf("  --title       Article title (default: %q)\n", defaultTitle)
	fmt.Printf("  --description Article description (default: %q)\n", defaultDescription)
	fmt.Printf("  --tags        Comma-separated tags (default: %q)\n", defaultTags)
	fmt.Printf("  --category    Article category (default: %q)\n", defaultCategory)
	fmt.Println("  --author      Author name (default: current OS username)")
	fmt.Printf("  --draft       Mark article as draft (default: %v)\n", defaultDraft)
	fmt.Printf("  --featured    Mark article as featured (default: %v)\n", defaultFeatured)
	fmt.Println("  --help        Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  new-article")
	fmt.Println("  new-article --title \"Hello World\" --tags \"golang,tutorial\"")
	fmt.Println("  new-article --title \"My Post\" --description \"Great post\" --draft=false --featured=true")
}

func slugify(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Define common stop words to remove
	stopWords := []string{"the", "a", "an", "in", "on", "at", "with", "for",
		"to", "of", "is", "are", "and", "or", "but", "from",
		"by", "as", "if", "when", "how", "what", "where", "this"}

	// Split into words and remove stop words (but keep first word if meaningful)
	words := strings.Fields(slug)
	var meaningful []string

	for i, word := range words {
		// Always keep first word, or keep if not a stop word
		if i == 0 || !isStopWord(word, stopWords) {
			// Clean the word of non-alphanumeric characters
			cleanWord := regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(word, "")
			if cleanWord != "" {
				meaningful = append(meaningful, cleanWord)
				if len(meaningful) >= 5 { // Limit to 5 words max
					break
				}
			}
		}
	}

	// Join words with hyphens
	if len(meaningful) == 0 {
		return "untitled"
	}

	return strings.Join(meaningful, "-")
}

func isStopWord(word string, stopWords []string) bool {
	for _, stopWord := range stopWords {
		if word == stopWord {
			return true
		}
	}
	return false
}

func generateArticle(title, description, tagsStr, category, author string, isDraft, isFeatured bool) string {
	now := time.Now().Format(time.RFC3339)

	// Format tags as YAML array
	tagList := strings.Split(tagsStr, ",")
	for i, tag := range tagList {
		tagList[i] = strings.TrimSpace(tag)
	}
	formattedTags := strings.Join(tagList, ", ")

	return fmt.Sprintf(`---
title: "%s"
description: "%s"
date: %s
tags: [%s]
categories: [%s]
featured: %v
draft: %v
author: %s
---

# %s

Content goes here...

## Introduction

Write your introduction here.

## Main Content

Add your main content sections here.

## Conclusion

Wrap up your article with a conclusion.

---

*Written by %s on %s*
`, title, description, now, formattedTags, category, isFeatured, isDraft, author, title, author, now)
}
