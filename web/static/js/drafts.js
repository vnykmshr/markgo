/**
 * Drafts Page Module — one-click publish from drafts list.
 * Loaded/unloaded by app.js based on data-template="drafts".
 */

import { showToast } from './modules/toast.js';
import { authenticatedJSON } from './modules/auth-fetch.js';

let ac = null;

function updateDraftCount() {
    const subtitle = document.querySelector('.page-subtitle');
    if (subtitle) {
        const remaining = document.querySelectorAll('.card-wrapper').length;
        subtitle.textContent = `${remaining} draft${remaining !== 1 ? 's' : ''}`;
    }
}

export function init() {
    ac = new AbortController();
    const { signal } = ac;

    document.querySelectorAll('.draft-publish-btn').forEach((btn) => {
        btn.addEventListener('click', () => handlePublish(btn), { signal });
    });
}

export function destroy() {
    ac?.abort();
    ac = null;
}

async function handlePublish(btn) {
    const slug = btn.dataset.slug;
    if (!slug) {
        showToast('Unable to publish \u2014 please reload the page', 'error');
        return;
    }

    btn.disabled = true;
    btn.textContent = 'Publishing\u2026';

    try {
        const result = await authenticatedJSON(`/compose/publish/${encodeURIComponent(slug)}`, {
            method: 'POST',
        });

        if (result.status === 401 || result.status === 403) {
            showToast('Please sign in to publish', 'warning');
            btn.disabled = false;
            btn.textContent = 'Publish';
            return;
        }

        if (!result.ok) {
            showToast(result.error || 'Publish failed', 'error');
            btn.disabled = false;
            btn.textContent = 'Publish';
            return;
        }

        // Remove card with fade, update count after removal
        const wrapper = btn.closest('.card-wrapper');
        if (wrapper) {
            let removed = false;
            const removeCard = () => {
                if (removed) return;
                removed = true;
                if (wrapper.parentNode) wrapper.remove();
                updateDraftCount();
            };
            wrapper.style.transition = 'opacity 0.3s ease, transform 0.3s ease';
            wrapper.style.opacity = '0';
            wrapper.style.transform = 'translateY(-8px)';
            wrapper.addEventListener('transitionend', removeCard, { once: true });
            setTimeout(removeCard, 400);
        }

        showToast(result.data.message || 'Published', 'success');
    } catch (err) {
        console.error('Draft publish failed:', err);
        showToast('Network error \u2014 please try again', 'error');
        btn.disabled = false;
        btn.textContent = 'Publish';
    }
}
