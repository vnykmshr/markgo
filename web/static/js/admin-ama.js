/**
 * Admin AMA — publish answers and delete spam.
 */

import { showToast } from './modules/toast.js';
import { authenticatedJSON } from './modules/auth-fetch.js';

let ac = null;

export function init() {
    ac = new AbortController();
    const { signal } = ac;

    document.querySelectorAll('.ama-publish-btn').forEach((btn) => {
        btn.addEventListener('click', () => publishAnswer(btn), { signal });
    });

    document.querySelectorAll('.ama-delete-btn').forEach((btn) => {
        btn.addEventListener('click', () => deleteQuestion(btn), { signal });
    });
}

async function publishAnswer(btn) {
    if (btn.disabled) return;

    const slug = btn.dataset.slug;
    const card = btn.closest('.ama-card');
    const textarea = card.querySelector('.ama-card-textarea');
    const answer = textarea.value.trim();

    if (!answer) {
        showToast('Write an answer first', 'warning');
        textarea.focus();
        return;
    }

    btn.disabled = true;
    btn.textContent = 'Publishing\u2026';

    try {
        const result = await authenticatedJSON(`/admin/ama/${slug}/answer`, {
            method: 'POST',
            body: { answer },
        });

        if (result.ok) {
            removeCard(card);
            showToast('Answer published', 'success');
        } else {
            showToast(result.error || 'Failed to publish', 'error');
            btn.disabled = false;
            btn.textContent = 'Publish';
        }
    } catch {
        showToast('Network error \u2014 try again', 'error');
        btn.disabled = false;
        btn.textContent = 'Publish';
    }
}

async function deleteQuestion(btn) {
    if (btn.disabled) return;

    const slug = btn.dataset.slug;
    if (!confirm('Delete this question? This cannot be undone.')) return;

    btn.disabled = true;
    btn.textContent = 'Deleting\u2026';

    const card = btn.closest('.ama-card');

    try {
        const result = await authenticatedJSON(`/admin/ama/${slug}/delete`, {
            method: 'POST',
        });

        if (result.ok) {
            removeCard(card);
            showToast('Question deleted', 'success');
        } else {
            showToast(result.error || 'Failed to delete', 'error');
            btn.disabled = false;
            btn.textContent = 'Delete';
        }
    } catch {
        showToast('Network error \u2014 try again', 'error');
        btn.disabled = false;
        btn.textContent = 'Delete';
    }
}

function removeCard(card) {
    const prefersReduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (prefersReduced) {
        card.remove();
        updateCount();
    } else {
        card.style.opacity = '0';
        card.style.transform = 'translateX(20px)';
        card.style.transition = 'opacity 0.3s, transform 0.3s';
        setTimeout(() => {
            card.remove();
            updateCount();
        }, 300);
    }
}

function updateCount() {
    const remaining = document.querySelectorAll('.ama-card').length;
    const subtitle = document.querySelector('.page-subtitle');
    if (subtitle) {
        subtitle.textContent = `${remaining} pending question${remaining !== 1 ? 's' : ''}`;
    }

    // Show empty state if no more cards
    if (remaining === 0) {
        const section = document.querySelector('.ama-pending');
        if (section) {
            const empty = document.createElement('div');
            empty.className = 'empty-state';

            const h2 = document.createElement('h2');
            h2.textContent = 'No pending questions';

            const p = document.createElement('p');
            p.textContent = 'When readers submit questions via the AMA form, they\u2019ll appear here for moderation.';

            empty.appendChild(h2);
            empty.appendChild(p);

            section.replaceChildren(empty);
        }
    }
}

export function destroy() {
    if (ac) { ac.abort(); ac = null; }
}
