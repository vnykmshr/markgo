package services

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apperrors "github.com/vnykmshr/markgo/internal/errors"
	"github.com/vnykmshr/markgo/internal/models"
)

func TestArticleService_DraftOperations(t *testing.T) {
	// Create a temporary directory for test articles
	tempDir := t.TempDir()

	// Create test articles - one draft, one published
	draftContent := `---
title: "Draft Article"
description: "This is a draft"
date: 2024-01-15T00:00:00Z
tags:
  - test
  - draft
categories:
  - testing
author: Test Author
draft: true
featured: false
---

# Draft Article

This is draft content that should not appear in regular listings.
`

	publishedContent := `---
title: "Published Article"
description: "This is published"
date: 2024-01-16T00:00:00Z
tags:
  - test
  - published
categories:
  - testing
author: Test Author
draft: false
featured: false
---

# Published Article

This is published content.
`

	// Write test files
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "draft-article.markdown"), []byte(draftContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "published-article.markdown"), []byte(publishedContent), 0644))

	// Create article service
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	service, err := NewArticleService(tempDir, logger)
	require.NoError(t, err)

	t.Run("GetDraftArticles", func(t *testing.T) {
		drafts := service.GetDraftArticles()

		assert.Len(t, drafts, 1)
		assert.Equal(t, "draft-article", drafts[0].Slug)
		assert.Equal(t, "Draft Article", drafts[0].Title)
		assert.True(t, drafts[0].Draft)
	})

	t.Run("GetDraftBySlug", func(t *testing.T) {
		// Test getting existing draft
		draft, err := service.GetDraftBySlug("draft-article")
		require.NoError(t, err)
		assert.Equal(t, "draft-article", draft.Slug)
		assert.True(t, draft.Draft)

		// Test getting published article (should also work)
		published, err := service.GetDraftBySlug("published-article")
		require.NoError(t, err)
		assert.Equal(t, "published-article", published.Slug)
		assert.False(t, published.Draft)

		// Test non-existent slug
		_, err = service.GetDraftBySlug("non-existent")
		assert.Error(t, err)
		assert.True(t, apperrors.IsArticleNotFound(err))
	})

	t.Run("PreviewDraft", func(t *testing.T) {
		// Test previewing draft
		preview, err := service.PreviewDraft("draft-article")
		require.NoError(t, err)
		assert.Equal(t, "draft-article", preview.Slug)
		assert.NotEmpty(t, preview.GetProcessedContent()) // Content should be processed
		assert.NotEmpty(t, preview.GetExcerpt())

		// Test non-existent draft
		_, err = service.PreviewDraft("non-existent")
		assert.Error(t, err)
		assert.True(t, apperrors.IsArticleNotFound(err))
	})

	t.Run("PublishDraft", func(t *testing.T) {
		// Verify draft is initially a draft
		draft, err := service.GetDraftBySlug("draft-article")
		require.NoError(t, err)
		assert.True(t, draft.Draft)

		// Publish the draft
		err = service.PublishDraft("draft-article")
		require.NoError(t, err)

		// Verify it's now published
		article, err := service.GetDraftBySlug("draft-article")
		require.NoError(t, err)
		assert.False(t, article.Draft)

		// Verify it appears in published articles list
		articles := service.GetAllArticles()
		found := false
		for _, a := range articles {
			if a.Slug == "draft-article" {
				found = true
				break
			}
		}
		assert.True(t, found, "Published draft should appear in articles list")

		// Try to publish again (should fail)
		err = service.PublishDraft("draft-article")
		assert.Error(t, err)
		assert.True(t, apperrors.IsValidationError(err))

		// Test non-existent draft
		err = service.PublishDraft("non-existent")
		assert.Error(t, err)
		assert.True(t, apperrors.IsArticleNotFound(err))
	})

	t.Run("UnpublishArticle", func(t *testing.T) {
		// Verify article is initially published (from previous test)
		article, err := service.GetDraftBySlug("draft-article")
		require.NoError(t, err)
		assert.False(t, article.Draft)

		// Unpublish the article
		err = service.UnpublishArticle("draft-article")
		require.NoError(t, err)

		// Verify it's now a draft
		draft, err := service.GetDraftBySlug("draft-article")
		require.NoError(t, err)
		assert.True(t, draft.Draft)

		// Verify it no longer appears in published articles list
		articles := service.GetAllArticles()
		found := false
		for _, a := range articles {
			if a.Slug == "draft-article" {
				found = true
				break
			}
		}
		assert.False(t, found, "Unpublished article should not appear in published articles list")

		// Try to unpublish again (should fail)
		err = service.UnpublishArticle("draft-article")
		assert.Error(t, err)
		assert.True(t, apperrors.IsValidationError(err))

		// Test non-existent article
		err = service.UnpublishArticle("non-existent")
		assert.Error(t, err)
		assert.True(t, apperrors.IsArticleNotFound(err))
	})

	t.Run("DraftsSortedByModified", func(t *testing.T) {
		drafts := service.GetDraftArticles()

		// Should have at least one draft (the unpublished one)
		assert.GreaterOrEqual(t, len(drafts), 1)

		// Should be sorted by last modified (most recent first)
		if len(drafts) > 1 {
			for i := 1; i < len(drafts); i++ {
				assert.True(t, drafts[i-1].LastModified.After(drafts[i].LastModified) ||
					drafts[i-1].LastModified.Equal(drafts[i].LastModified))
			}
		}
	})
}

func TestArticleService_DraftFileOperations(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test article file
	content := `---
title: "File Operation Test"
description: "Testing file operations"
date: 2024-01-17T00:00:00Z
tags: [test]
categories: [testing]
author: Test Author
draft: true
featured: false
---

# File Operation Test

Testing file operations.
`

	filePath := filepath.Join(tempDir, "file-operation-test.markdown")
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))

	t.Run("updateArticleFile", func(t *testing.T) {
		service, err := NewArticleService(tempDir, logger)
		require.NoError(t, err)

		// Get the article
		article, err := service.GetDraftBySlug("file-operation-test")
		require.NoError(t, err)
		assert.True(t, article.Draft)

		// Update the article's draft status in memory
		article.Draft = false

		// Update the file
		err = service.updateArticleFile("file-operation-test", article)
		require.NoError(t, err)

		// Verify file was updated by reading it back
		updatedContent, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Contains(t, string(updatedContent), "draft: false")
		assert.NotContains(t, string(updatedContent), "draft: true")
	})

	t.Run("rebuildArticlesList", func(t *testing.T) {
		service, err := NewArticleService(tempDir, logger)
		require.NoError(t, err)

		// Get initial count (should be 0 since the article in tempDir is a draft)
		initialPublished := len(service.GetAllArticles())

		// Manually add an article to cache and mark as published
		testArticle := &models.Article{
			Slug:  "rebuild-test",
			Title: "Rebuild Test",
			Draft: false,
		}
		service.cache["rebuild-test"] = testArticle

		// Rebuild articles list
		service.rebuildArticlesList()

		// Should now include the new article
		finalPublished := len(service.GetAllArticles())
		assert.Equal(t, initialPublished+1, finalPublished)

		// Clean up
		delete(service.cache, "rebuild-test")
		service.rebuildArticlesList()
	})
}
