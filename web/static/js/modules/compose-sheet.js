/**
 * Compose Sheet — quick capture overlay.
 *
 * Shell module: runs once, persists across SPA navigations.
 * Opens on fab:compose event, closes on backdrop/Escape/navigation.
 * Submits to POST /compose/quick with JSON + X-CSRF-Token header.
 */

import { showToast } from './toast.js';

let overlay = null;
let textarea = null;
let titleInput = null;
let publishBtn = null;
let isOpen = false;
let isSubmitting = false;

function getCSRFToken() {
    return document.querySelector('meta[name="csrf-token"]')?.content || '';
}

function buildOverlay() {
    const el = document.createElement('div');
    el.className = 'compose-sheet-overlay';
    el.setAttribute('role', 'dialog');
    el.setAttribute('aria-modal', 'true');
    el.setAttribute('aria-label', 'Quick compose');
    el.hidden = true;

    // Backdrop
    const backdrop = document.createElement('div');
    backdrop.className = 'compose-sheet-backdrop';
    backdrop.addEventListener('click', close);

    // Sheet
    const sheet = document.createElement('div');
    sheet.className = 'compose-sheet';

    // Header
    const header = document.createElement('div');
    header.className = 'compose-sheet-header';

    const heading = document.createElement('span');
    heading.className = 'compose-sheet-heading';
    heading.textContent = 'Quick compose';

    const closeBtn = document.createElement('button');
    closeBtn.className = 'compose-sheet-close';
    closeBtn.setAttribute('aria-label', 'Close');
    closeBtn.textContent = '\u00D7';
    closeBtn.addEventListener('click', close);

    header.appendChild(heading);
    header.appendChild(closeBtn);

    // Title toggle + input
    const titleGroup = document.createElement('div');
    titleGroup.className = 'compose-sheet-title-group';

    const titleToggle = document.createElement('button');
    titleToggle.className = 'compose-sheet-title-toggle';
    titleToggle.type = 'button';
    titleToggle.textContent = '+ Add title';
    titleToggle.setAttribute('aria-expanded', 'false');
    titleToggle.setAttribute('aria-controls', 'compose-sheet-title-input');
    titleToggle.addEventListener('click', () => {
        titleGroup.classList.toggle('expanded');
        if (titleGroup.classList.contains('expanded')) {
            titleToggle.textContent = '\u2212 Remove title';
            titleToggle.setAttribute('aria-expanded', 'true');
            titleInput.focus();
        } else {
            titleToggle.textContent = '+ Add title';
            titleToggle.setAttribute('aria-expanded', 'false');
            titleInput.value = '';
        }
    });

    titleInput = document.createElement('input');
    titleInput.type = 'text';
    titleInput.id = 'compose-sheet-title-input';
    titleInput.className = 'compose-sheet-title-input';
    titleInput.placeholder = 'Title (optional)';
    titleInput.maxLength = 200;

    titleGroup.appendChild(titleToggle);
    titleGroup.appendChild(titleInput);

    // Textarea
    textarea = document.createElement('textarea');
    textarea.className = 'compose-sheet-textarea';
    textarea.placeholder = 'What\u2019s on your mind?';
    textarea.rows = 4;

    // Footer
    const footer = document.createElement('div');
    footer.className = 'compose-sheet-footer';

    const wordCount = document.createElement('span');
    wordCount.className = 'compose-sheet-wordcount';
    wordCount.textContent = 'thought';

    textarea.addEventListener('input', () => {
        const words = textarea.value.trim().split(/\s+/).filter(Boolean).length;
        const hasTitle = titleGroup.classList.contains('expanded') && titleInput.value.trim();
        if (hasTitle) {
            wordCount.textContent = 'article';
        } else if (words < 100) {
            wordCount.textContent = 'thought';
        } else {
            wordCount.textContent = 'article';
        }
    });

    publishBtn = document.createElement('button');
    publishBtn.className = 'compose-sheet-publish';
    publishBtn.textContent = 'Publish';
    publishBtn.addEventListener('click', handlePublish);

    footer.appendChild(wordCount);
    footer.appendChild(publishBtn);

    // Assemble
    sheet.appendChild(header);
    sheet.appendChild(titleGroup);
    sheet.appendChild(textarea);
    sheet.appendChild(footer);

    el.appendChild(backdrop);
    el.appendChild(sheet);

    return el;
}

function open() {
    if (isOpen) return;
    if (!overlay) {
        overlay = buildOverlay();
        document.body.appendChild(overlay);
    }

    overlay.hidden = false;
    isOpen = true;

    // Prevent body scroll
    document.body.style.overflow = 'hidden';

    // Focus textarea after animation
    requestAnimationFrame(() => {
        textarea.focus();
    });

    // Escape key
    document.addEventListener('keydown', handleKeydown);
}

function close() {
    if (!isOpen) return;
    isOpen = false;

    overlay.hidden = true;
    document.body.style.overflow = '';
    document.removeEventListener('keydown', handleKeydown);

    // Reset form
    textarea.value = '';
    titleInput.value = '';
    const titleGroup = overlay.querySelector('.compose-sheet-title-group');
    if (titleGroup) {
        titleGroup.classList.remove('expanded');
        const toggle = titleGroup.querySelector('.compose-sheet-title-toggle');
        if (toggle) {
            toggle.textContent = '+ Add title';
            toggle.setAttribute('aria-expanded', 'false');
        }
    }

    // Reset publish button
    publishBtn.disabled = false;
    publishBtn.textContent = 'Publish';
    isSubmitting = false;
}

function handleKeydown(e) {
    if (e.key === 'Escape') {
        e.preventDefault();
        close();
    }
}

async function handlePublish() {
    if (isSubmitting) return;

    const content = textarea.value.trim();
    if (!content) {
        showToast('Write something first', 'warning');
        textarea.focus();
        return;
    }

    const token = getCSRFToken();
    if (!token) {
        showToast('Session expired — please refresh', 'error');
        return;
    }

    isSubmitting = true;
    publishBtn.disabled = true;
    publishBtn.textContent = 'Publishing\u2026';

    const titleGroup = overlay.querySelector('.compose-sheet-title-group');
    const hasTitle = titleGroup?.classList.contains('expanded');
    const title = hasTitle ? titleInput.value.trim() : '';

    const body = { content };
    if (title) body.title = title;

    try {
        const response = await fetch('/compose/quick', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': token,
            },
            body: JSON.stringify(body),
        });

        const data = await response.json();

        if (response.ok) {
            close();
            showToast(data.message || 'Published!', 'success');
        } else if (response.status === 401) {
            showToast('Not authenticated — please sign in', 'error');
            isSubmitting = false;
            publishBtn.disabled = false;
            publishBtn.textContent = 'Publish';
        } else {
            showToast(data.error || 'Failed to publish', 'error');
            isSubmitting = false;
            publishBtn.disabled = false;
            publishBtn.textContent = 'Publish';
        }
    } catch {
        showToast('Network error — try again', 'error');
        isSubmitting = false;
        publishBtn.disabled = false;
        publishBtn.textContent = 'Publish';
    }
}

export function init() {
    // Listen for FAB trigger
    document.addEventListener('fab:compose', open);

    // Close on SPA navigation
    document.addEventListener('router:navigate-start', close);
}
