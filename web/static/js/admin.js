/**
 * Admin Dashboard — POST action execution with toast feedback.
 */

import { showToast } from './modules/toast.js';

let ac = null;

export function init() {
    ac = new AbortController();
    const { signal } = ac;

    document.querySelectorAll('.admin-action-btn').forEach((button) => {
        button.addEventListener('click', () => executeAction(button), { signal });
    });
}

async function executeAction(button) {
    if (button.disabled) return;

    const url = button.dataset.url;
    const label = button.dataset.label || button.textContent;

    if (!confirm(`Are you sure you want to ${label.toLowerCase()}?`)) return;

    const originalText = button.textContent;
    button.disabled = true;
    button.textContent = 'Working...';

    try {
        const response = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
        });

        if (!response.ok) {
            let msg = `${label} failed (HTTP ${response.status})`;
            try {
                const body = await response.json();
                if (body.error) msg = body.error;
            } catch { /* response not JSON, use default */ }
            throw new Error(msg);
        }

        const data = await response.json().catch(() => ({}));
        showToast(data.message || `${label} completed`, 'success');
    } catch (error) {
        const msg = error instanceof TypeError
            ? `${label} failed — check your network connection`
            : error.message || `${label} failed`;
        showToast(msg, 'error');
    } finally {
        button.textContent = originalText;
        button.disabled = false;
    }
}

export function destroy() {
    if (ac) { ac.abort(); ac = null; }
}
