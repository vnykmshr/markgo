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

    // =========================================================================
    // Image Upload
    // =========================================================================

    var uploadBtn = document.querySelector('.compose-upload-btn');
    var fileInput = document.querySelector('.compose-file-input');
    var uploadStatus = document.querySelector('.upload-status');

    var allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    var maxFileSize = 5 * 1024 * 1024; // 5MB

    function insertAtCursor(text) {
        var start = textarea.selectionStart;
        var end = textarea.selectionEnd;
        var before = textarea.value.substring(0, start);
        var after = textarea.value.substring(end);
        textarea.value = before + text + after;
        textarea.selectionStart = textarea.selectionEnd = start + text.length;
        textarea.focus();
        if (previewVisible) debouncedPreview();
    }

    function setUploadStatus(msg, isError) {
        if (!uploadStatus) return;
        uploadStatus.textContent = msg;
        uploadStatus.className = 'upload-status' + (isError ? ' upload-error' : msg ? ' upload-success' : '');
    }

    function uploadFile(file) {
        if (!file) return;

        if (allowedTypes.indexOf(file.type) === -1) {
            setUploadStatus('Only JPEG, PNG, GIF, and WebP images are allowed.', true);
            return;
        }
        if (file.size > maxFileSize) {
            setUploadStatus('File too large. Maximum size is 5MB.', true);
            return;
        }

        setUploadStatus('Uploading...', false);

        var formData = new FormData();
        formData.append('file', file);
        if (csrfInput) formData.append('_csrf', csrfInput.value);

        fetch('/compose/upload', {
            method: 'POST',
            body: formData,
            credentials: 'same-origin'
        })
        .then(function (res) {
            if (!res.ok) return res.json().then(function (data) { throw new Error(data.error || 'Upload failed'); });
            return res.json();
        })
        .then(function (data) {
            insertAtCursor(data.markdown + '\n');
            setUploadStatus('Uploaded: ' + data.filename, false);
        })
        .catch(function (err) {
            setUploadStatus(err.message || 'Upload failed', true);
        });
    }

    if (uploadBtn && fileInput) {
        uploadBtn.addEventListener('click', function () {
            fileInput.click();
        });

        fileInput.addEventListener('change', function () {
            if (fileInput.files.length > 0) {
                uploadFile(fileInput.files[0]);
                fileInput.value = '';
            }
        });
    }

    // Drag and drop on textarea
    textarea.addEventListener('dragover', function (e) {
        e.preventDefault();
        textarea.classList.add('compose-textarea-dragover');
    });

    textarea.addEventListener('dragleave', function () {
        textarea.classList.remove('compose-textarea-dragover');
    });

    textarea.addEventListener('drop', function (e) {
        e.preventDefault();
        textarea.classList.remove('compose-textarea-dragover');
        var files = e.dataTransfer.files;
        if (files.length > 0) uploadFile(files[0]);
    });
})();
