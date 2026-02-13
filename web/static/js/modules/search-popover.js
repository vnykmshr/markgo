/**
 * Search Popover — header search with Cmd/Ctrl+K shortcut.
 *
 * Shell module: persists across SPA navigations.
 * Form submit navigates via SPA router to /search?q=...
 */

import { initPopover } from './popover.js';

let popoverCtrl = null;

export function init() {
    popoverCtrl = initPopover('search-popover', '.search-trigger', (popover) => {
        const input = popover.querySelector('.search-popover-input');
        if (input) input.focus();
    }, (popover) => {
        const input = popover.querySelector('.search-popover-input');
        if (input) input.value = '';
    });

    // Form submit → SPA navigate
    const form = document.getElementById('search-popover-form');
    if (form) {
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            const input = form.querySelector('.search-popover-input');
            const query = input?.value.trim();
            if (!query) return;

            if (popoverCtrl) popoverCtrl.close();

            // Navigate via SPA router — synthetic link click triggers router interception
            const a = document.createElement('a');
            a.href = '/search?q=' + encodeURIComponent(query);
            a.style.display = 'none';
            document.body.appendChild(a);
            a.click();
            a.remove();
        });
    }

    // Cmd/Ctrl+K shortcut
    document.addEventListener('keydown', (e) => {
        if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
            e.preventDefault();
            if (!popoverCtrl) return;
            if (popoverCtrl.popover.hidden) {
                popoverCtrl.open();
            } else {
                popoverCtrl.close();
            }
        }
    });

    // Close on SPA navigation
    document.addEventListener('router:navigate-start', () => {
        if (popoverCtrl) popoverCtrl.close();
    });
}
