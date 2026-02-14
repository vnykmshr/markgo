/**
 * Admin Dashboard — POST action execution with toast feedback.
 */

import { showToast } from './modules/toast.js';
import { authenticatedJSON } from './modules/auth-fetch.js';

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
        const result = await authenticatedJSON(url, { method: 'POST' });

        if (!result.ok) {
            showToast(result.error || `${label} failed`, 'error');
        } else {
            showToast(result.data.message || `${label} completed`, 'success');
        }
    } catch (error) {
        showToast(`${label} failed \u2014 check your network connection`, 'error');
    } finally {
        button.textContent = originalText;
        button.disabled = false;
    }
}

export function destroy() {
    if (ac) { ac.abort(); ac = null; }
}
