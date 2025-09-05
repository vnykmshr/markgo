/**
 * Article Page JavaScript
 * Handles article-specific functionality like link copying and syntax highlighting
 * @fileoverview Article page functionality for MarkGo blog engine
 */

/**
 * Copies text to the user's clipboard and provides visual feedback
 * @param {string} text - The text content to copy to clipboard
 * @global
 */
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(function() {
        // Show success message
        const button = event.target.closest('.copy-link');
        const originalText = button.innerHTML;
        button.innerHTML = '<svg width="20" height="20" fill="currentColor" viewBox="0 0 24 24"><path d="M9 16.2L4.8 12l-1.4 1.4L9 19 21 7l-1.4-1.4L9 16.2z"/></svg>Copied!';
        button.style.backgroundColor = 'var(--color-success)';
        button.style.color = 'white';
        button.style.borderColor = 'var(--color-success)';

        setTimeout(function() {
            button.innerHTML = originalText;
            button.style.backgroundColor = '';
            button.style.color = '';
            button.style.borderColor = '';
        }, 2000);
    }).catch(function(err) {
        console.error('Failed to copy: ', err);
    });
}

/**
 * Initialize article page functionality
 * Sets up syntax highlighting for code blocks
 */
document.addEventListener('DOMContentLoaded', function() {
    // Initialize syntax highlighting for all code blocks
    if (typeof hljs !== 'undefined') {
        hljs.highlightAll();
    }
});
