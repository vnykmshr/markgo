(function () {
    'use strict';

    var textarea = document.getElementById('content');
    var previewBtn = document.querySelector('.compose-preview-btn');
    var previewPanel = document.querySelector('.compose-preview-panel');
    var previewContent = document.querySelector('.compose-preview-content');
    var csrfInput = document.querySelector('input[name="_csrf"]');
    var form = document.querySelector('.compose-form');

    if (!textarea || !previewBtn || !previewPanel || !previewContent) return;

    var previewVisible = false;
    var debounceTimer = null;

    // NOTE: innerHTML is used intentionally here to render server-rendered HTML
    // from our markdown renderer. This is behind BasicAuth (admin-only), so
    // self-XSS is an accepted known limitation — same as the article page.
    function setPreviewHTML(html) {
        previewContent.innerHTML = html; // nosemgrep: javascript.browser.security.innerHTML
    }

    function fetchPreview() {
        var content = textarea.value;
        if (!content.trim()) {
            setPreviewHTML('<p class="preview-placeholder">Start typing to see a preview...</p>');
            return;
        }

        setPreviewHTML('<p class="preview-loading">Rendering preview...</p>');

        var formData = new FormData();
        formData.append('content', content);
        if (csrfInput) formData.append('_csrf', csrfInput.value);

        fetch('/compose/preview', {
            method: 'POST',
            body: formData,
            credentials: 'same-origin'
        })
        .then(function (res) {
            if (!res.ok) throw new Error('Preview failed');
            return res.text();
        })
        .then(function (html) {
            setPreviewHTML(html);
            if (typeof hljs !== 'undefined') {
                previewContent.querySelectorAll('pre code').forEach(function (block) {
                    hljs.highlightElement(block);
                });
            }
        })
        .catch(function () {
            setPreviewHTML('<p class="preview-error">Preview unavailable</p>');
        });
    }

    function debouncedPreview() {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(fetchPreview, 500);
    }

    previewBtn.addEventListener('click', function () {
        previewVisible = !previewVisible;
        form.classList.toggle('preview-visible', previewVisible);
        previewBtn.textContent = previewVisible ? 'Edit' : 'Preview';
        previewBtn.setAttribute('aria-pressed', previewVisible);

        if (previewVisible) {
            fetchPreview();
            textarea.addEventListener('input', debouncedPreview);
        } else {
            textarea.removeEventListener('input', debouncedPreview);
            clearTimeout(debounceTimer);
        }
    });
})();
