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

function executeAction(button) {
    const url = button.dataset.url;
    const label = button.dataset.label || button.textContent;
    const originalText = button.textContent;

    button.textContent = 'Working...';
    button.disabled = true;

    fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    })
        .then((response) => {
            if (!response.ok) throw new Error(`HTTP ${response.status}`);
            return response.json();
        })
        .then((data) => {
            showToast(data.message || `${label} completed`, 'success');
        })
        .catch((error) => {
            showToast(error.message || `${label} failed`, 'error');
        })
        .finally(() => {
            button.textContent = originalText;
            button.disabled = false;
        });
}

export function destroy() {
    if (ac) { ac.abort(); ac = null; }
}
