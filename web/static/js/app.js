/**
 * MarkGo Engine — ES Module Entry Point
 *
 * Shell modules (navigation, theme, scroll, login) run once and persist.
 * Content modules (highlight, lazy) re-run after each SPA navigation.
 * Page-specific modules load/unload based on data-template attribute.
 * Router intercepts links and swaps <main> content without full reloads.
 */

import { init as initNavigation } from './modules/navigation.js';
import { init as initTheme } from './modules/theme.js';
import { init as initHighlight } from './modules/highlight.js';
import { init as initScroll } from './modules/scroll.js';
import { init as initLazy } from './modules/lazy.js';
import { init as initLogin } from './modules/login.js';
import { init as initToast, showToast } from './modules/toast.js';
import { init as initFab } from './modules/fab.js';
import { init as initComposeSheet } from './modules/compose-sheet.js';
import { init as initSearchPopover } from './modules/search-popover.js';
import { init as initSubscribePopover } from './modules/subscribe-popover.js';
import { init as initRouter } from './modules/router.js';

// Page-specific module loaders
const PAGE_MODULES = {
    search: () => import('./search-page.js'),
    about: () => import('./contact.js'),
    compose: () => import('./compose.js'),
    admin_home: () => import('./admin.js'),
    drafts: () => import('./drafts.js'),
};

let currentPageModule = null;

async function loadPageModule(template) {
    // Cleanup previous module if it supports it
    if (currentPageModule?.destroy) currentPageModule.destroy();
    currentPageModule = null;

    const loader = PAGE_MODULES[template];
    if (loader) {
        try {
            const mod = await loader();
            mod.init();
            currentPageModule = mod;
        } catch (err) {
            console.error(`Failed to load page module for "${template}":`, err);
        }
    }
}

/**
 * Called by the router after content swap.
 * Re-runs content-dependent modules and loads page-specific JS.
 */
function reinitPage(template) {
    initHighlight();
    initLazy();
    loadPageModule(template);
}

// ── Service Worker registration ──────────────────────────────────────────────

if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/sw.js').catch((err) => {
        console.error('Service worker registration failed:', err);
    });
}

// ── Install prompt ───────────────────────────────────────────────────────────

const INSTALL_VISIT_KEY = 'markgo:visit-count';
const INSTALL_DISMISSED_KEY = 'markgo:install-dismissed';
const INSTALL_THRESHOLD = 3;
const INSTALL_DISMISS_DAYS = 30;

let deferredPrompt = null;

window.addEventListener('beforeinstallprompt', (e) => {
    e.preventDefault();
    deferredPrompt = e;

    // Don't show if user dismissed recently (re-prompt after 30 days)
    try {
        const dismissedAt = parseInt(localStorage.getItem(INSTALL_DISMISSED_KEY) || '0', 10);
        if (dismissedAt && Date.now() - dismissedAt < INSTALL_DISMISS_DAYS * 86400000) return;
    } catch { /* ignore */ }

    // Track visit count
    let visits = 0;
    try {
        visits = parseInt(localStorage.getItem(INSTALL_VISIT_KEY) || '0', 10) + 1;
        localStorage.setItem(INSTALL_VISIT_KEY, String(visits));
    } catch { /* ignore */ }

    if (visits >= INSTALL_THRESHOLD) {
        showInstallBanner();
    }
});

function showInstallBanner() {
    if (!deferredPrompt) return;

    const banner = document.createElement('div');
    banner.className = 'install-banner';
    banner.setAttribute('role', 'complementary');
    banner.setAttribute('aria-label', 'Install app');

    const text = document.createElement('span');
    text.className = 'install-banner-text';
    text.textContent = 'Install MarkGo for quick access';

    const installBtn = document.createElement('button');
    installBtn.className = 'install-banner-btn';
    installBtn.textContent = 'Install';
    installBtn.addEventListener('click', async () => {
        banner.remove();
        if (!deferredPrompt) return;
        deferredPrompt.prompt();
        const { outcome } = await deferredPrompt.userChoice;
        if (outcome === 'accepted') {
            showToast('App installed!', 'success');
        }
        deferredPrompt = null;
    });

    const dismissBtn = document.createElement('button');
    dismissBtn.className = 'install-banner-dismiss';
    dismissBtn.setAttribute('aria-label', 'Dismiss');
    dismissBtn.textContent = '\u00d7';
    dismissBtn.addEventListener('click', () => {
        banner.remove();
        try { localStorage.setItem(INSTALL_DISMISSED_KEY, String(Date.now())); } catch { /* ignore */ }
    });

    banner.appendChild(text);
    banner.appendChild(installBtn);
    banner.appendChild(dismissBtn);
    document.body.appendChild(banner);
}

// ── Offline indicator ────────────────────────────────────────────────────────

let offlineToast = null;
window.addEventListener('offline', () => {
    offlineToast = showToast('You are offline', 'warning', { duration: 0 });
});
window.addEventListener('online', () => {
    if (offlineToast) {
        offlineToast.dismiss();
        offlineToast = null;
    }
    showToast('Back online', 'success');
});

document.addEventListener('DOMContentLoaded', () => {
    // Shell modules — run once, persist across navigations
    initNavigation();
    initTheme();
    initScroll();
    initLogin();
    initToast();
    initFab();
    initComposeSheet();
    initSearchPopover();
    initSubscribePopover();

    // Content modules — initial page
    initHighlight();
    initLazy();

    // Page-specific module — initial page
    loadPageModule(document.body.dataset.template);

    // Router — last, passes reinitPage as callback
    initRouter(reinitPage);
});
