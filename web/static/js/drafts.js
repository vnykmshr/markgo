/**
 * Drafts Page Module — one-click publish from drafts list.
 * Loaded/unloaded by app.js based on data-template="drafts".
 */

import { showToast } from './modules/toast.js';

let controller = null;

function getCSRFToken() {
    return document.querySelector('meta[name="csrf-token"]')?.content || '';
}

export function init() {
    controller = new AbortController();
    const { signal } = controller;

    document.querySelectorAll('.draft-publish-btn').forEach((btn) => {
        btn.addEventListener('click', () => handlePublish(btn), { signal });
    });
}

export function destroy() {
    controller?.abort();
    controller = null;
}

async function handlePublish(btn) {
    const slug = btn.dataset.slug;
    if (!slug) return;

    btn.disabled = true;
    btn.textContent = 'Publishing\u2026';

    try {
        const res = await fetch(`/compose/publish/${encodeURIComponent(slug)}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': getCSRFToken(),
            },
        });

        if (res.status === 401 || res.status === 403) {
            showToast('Please sign in to publish', 'warning');
            document.querySelector('.login-trigger')?.click();
            btn.disabled = false;
            btn.textContent = 'Publish';
            return;
        }

        const data = await res.json();

        if (!res.ok) {
            showToast(data.error || 'Publish failed', 'error');
            btn.disabled = false;
            btn.textContent = 'Publish';
            return;
        }

        // Remove card with fade
        const wrapper = btn.closest('.draft-card-wrapper');
        if (wrapper) {
            wrapper.style.transition = 'opacity 0.3s ease, transform 0.3s ease';
            wrapper.style.opacity = '0';
            wrapper.style.transform = 'translateY(-8px)';
            wrapper.addEventListener('transitionend', () => wrapper.remove(), { once: true });
        }

        // Update draft count in subtitle
        const subtitle = document.querySelector('.page-subtitle');
        if (subtitle) {
            const remaining = document.querySelectorAll('.draft-card-wrapper').length - 1;
            subtitle.textContent = `${remaining} draft${remaining !== 1 ? 's' : ''}`;
        }

        showToast(data.message || 'Published', 'success');
    } catch {
        showToast('Network error — please try again', 'error');
        btn.disabled = false;
        btn.textContent = 'Publish';
    }
}
