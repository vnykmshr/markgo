/**
 * Drafts Page Module — one-click publish from drafts list.
 * Loaded/unloaded by app.js based on data-template="drafts".
 */

import { showToast } from './modules/toast.js';

let ac = null;

function getCSRFToken() {
    return document.querySelector('meta[name="csrf-token"]')?.content || '';
}

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
        showToast('Unable to publish — please reload the page', 'error');
        return;
    }

    const csrfToken = getCSRFToken();
    if (!csrfToken) {
        showToast('Session expired — please reload the page', 'warning');
        return;
    }

    btn.disabled = true;
    btn.textContent = 'Publishing\u2026';

    try {
        const res = await fetch(`/compose/publish/${encodeURIComponent(slug)}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken,
            },
        });

        if (res.status === 401 || res.status === 403) {
            showToast('Please sign in to publish', 'warning');
            document.dispatchEvent(new CustomEvent('auth:open-login'));
            btn.disabled = false;
            btn.textContent = 'Publish';
            return;
        }

        let data;
        try {
            data = await res.json();
        } catch (err) {
            console.error('Failed to parse publish response:', err);
            showToast(`Publish failed (${res.status})`, 'error');
            btn.disabled = false;
            btn.textContent = 'Publish';
            return;
        }

        if (!res.ok) {
            showToast(data.error || 'Publish failed', 'error');
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
            setTimeout(removeCard, 400); // Safety: remove even if transitionend never fires
        }

        showToast(data.message || 'Published', 'success');
    } catch (err) {
        console.error('Draft publish failed:', err);
        showToast('Network error — please try again', 'error');
        btn.disabled = false;
        btn.textContent = 'Publish';
    }
}
