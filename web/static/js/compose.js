/**
 * Compose Page — markdown preview (debounced) and image upload with drag-and-drop.
 */

export function init() {
    const textarea = document.getElementById('content');
    const previewBtn = document.querySelector('.compose-preview-btn');
    const previewPanel = document.querySelector('.compose-preview-panel');
    const previewContent = document.querySelector('.compose-preview-content');
    const csrfInput = document.querySelector('input[name="_csrf"]');
    const form = document.querySelector('.compose-form');

    if (!textarea || !previewBtn || !previewPanel || !previewContent) return;

    let previewVisible = false;
    let debounceTimer = null;

    /**
     * Safely set preview HTML from server-rendered markdown.
     * Uses DOMParser instead of innerHTML — admin-only, behind session auth.
     */
    function setPreviewHTML(html) {
        const doc = new DOMParser().parseFromString(html, 'text/html');
        previewContent.textContent = '';
        while (doc.body.firstChild) {
            previewContent.appendChild(doc.body.firstChild);
        }
    }

    function fetchPreview() {
        const content = textarea.value;
        if (!content.trim()) {
            setPreviewHTML('<p class="preview-placeholder">Start typing to see a preview...</p>');
            return;
        }

        setPreviewHTML('<p class="preview-loading">Rendering preview...</p>');

        const formData = new FormData();
        formData.append('content', content);
        if (csrfInput) formData.append('_csrf', csrfInput.value);

        fetch('/compose/preview', {
            method: 'POST',
            body: formData,
            credentials: 'same-origin',
        })
            .then((res) => {
                if (!res.ok) throw new Error('Preview failed');
                return res.text();
            })
            .then((html) => {
                setPreviewHTML(html);
                if (typeof hljs !== 'undefined') {
                    previewContent.querySelectorAll('pre code').forEach((block) => {
                        hljs.highlightElement(block);
                    });
                }
            })
            .catch((err) => {
                setPreviewHTML('<p class="preview-error">Preview unavailable' + (err.message ? ': ' + err.message : '') + '</p>');
            });
    }

    function debouncedPreview() {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(fetchPreview, 500);
    }

    previewBtn.addEventListener('click', () => {
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

    const uploadBtn = document.querySelector('.compose-upload-btn');
    const fileInput = document.querySelector('.compose-file-input');
    const uploadStatus = document.querySelector('.upload-status');

    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    const maxFileSize = 5 * 1024 * 1024; // 5MB

    function insertAtCursor(text) {
        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const before = textarea.value.substring(0, start);
        const after = textarea.value.substring(end);
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

        if (!allowedTypes.includes(file.type)) {
            setUploadStatus('Only JPEG, PNG, GIF, and WebP images are allowed.', true);
            return;
        }
        if (file.size > maxFileSize) {
            setUploadStatus('File too large. Maximum size is 5MB.', true);
            return;
        }

        setUploadStatus('Uploading...', false);

        const formData = new FormData();
        formData.append('file', file);
        if (csrfInput) formData.append('_csrf', csrfInput.value);

        fetch('/compose/upload', {
            method: 'POST',
            body: formData,
            credentials: 'same-origin',
        })
            .then((res) => {
                if (!res.ok) {
                    return res.json()
                        .catch(() => { throw new Error('Upload failed (server error)'); })
                        .then((data) => { throw new Error(data.error || 'Upload failed'); });
                }
                return res.json();
            })
            .then((data) => {
                insertAtCursor(data.markdown + '\n');
                setUploadStatus('Uploaded: ' + data.filename, false);
            })
            .catch((err) => {
                setUploadStatus(err.message || 'Upload failed', true);
            });
    }

    if (uploadBtn && fileInput) {
        uploadBtn.addEventListener('click', () => fileInput.click());
        fileInput.addEventListener('change', () => {
            if (fileInput.files.length > 0) {
                uploadFile(fileInput.files[0]);
                fileInput.value = '';
            }
        });
    }

    // Drag and drop on textarea
    textarea.addEventListener('dragover', (e) => {
        e.preventDefault();
        textarea.classList.add('compose-textarea-dragover');
    });

    textarea.addEventListener('dragleave', () => {
        textarea.classList.remove('compose-textarea-dragover');
    });

    textarea.addEventListener('drop', (e) => {
        e.preventDefault();
        textarea.classList.remove('compose-textarea-dragover');
        const files = e.dataTransfer.files;
        if (files.length > 0) uploadFile(files[0]);
    });
}
