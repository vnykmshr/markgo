/**
 * Subscribe Popover — header RSS/feed options.
 *
 * Shell module: persists across SPA navigations.
 * Copy buttons use Clipboard API for feed URLs.
 * Footer "Subscribe" link intercepted to open popover.
 */

import { initPopover } from './popover.js';
import { showToast } from './toast.js';

let popoverCtrl = null;

function copyToClipboard(text) {
    if (!navigator.clipboard?.writeText) {
        showToast('Copy not available \u2014 use the address bar instead', 'warning');
        return Promise.resolve();
    }
    return navigator.clipboard.writeText(text).then(() => {
        showToast('URL copied!', 'success');
    }).catch((err) => {
        console.warn('Clipboard write failed:', err.message);
        showToast('Copy failed \u2014 use the address bar instead', 'error');
    });
}

export function init() {
    popoverCtrl = initPopover('subscribe-popover', '.subscribe-trigger');

    // Copy buttons
    const rssBtn = document.getElementById('copy-rss-url');
    const jsonBtn = document.getElementById('copy-json-url');

    if (rssBtn) {
        rssBtn.addEventListener('click', () => {
            copyToClipboard(window.location.origin + '/feed.xml');
        });
    }

    if (jsonBtn) {
        jsonBtn.addEventListener('click', () => {
            copyToClipboard(window.location.origin + '/feed.json');
        });
    }

    // Bottom nav subscribe button — also opens popover (separate from header trigger)
    bindBottomNavSubscribe();

    // Footer "Subscribe" link → open popover
    const footerSubscribe = document.querySelector('.footer-subscribe');
    if (footerSubscribe && popoverCtrl) {
        footerSubscribe.addEventListener('click', (e) => {
            e.preventDefault();
            window.scrollTo({ top: 0, behavior: 'smooth' });
            // Small delay so scroll completes before popover opens
            setTimeout(() => popoverCtrl.open(), 300);
        });
    }

    // Close on SPA navigation
    document.addEventListener('router:navigate-start', () => {
        if (popoverCtrl) popoverCtrl.close();
    });
}

/**
 * Bind the bottom nav subscribe button to open the header popover.
 * Separate from popover.js because bottom-nav trigger is outside the navbar container.
 */
function bindBottomNavSubscribe() {
    const btn = document.querySelector('.bottom-nav-subscribe');
    if (!btn || !popoverCtrl) return;

    btn.addEventListener('click', (e) => {
        e.stopPropagation();
        popoverCtrl.open();
        // Scroll to top so the popover is visible
        window.scrollTo({ top: 0, behavior: 'smooth' });
    });
}
