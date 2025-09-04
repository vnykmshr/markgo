/* Article Page JavaScript */

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

// Initialize syntax highlighting
document.addEventListener('DOMContentLoaded', function() {
    hljs.highlightAll();
});
