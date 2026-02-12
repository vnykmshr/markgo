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
import { init as initToast } from './modules/toast.js';
import { init as initFab } from './modules/fab.js';
import { init as initComposeSheet } from './modules/compose-sheet.js';
import { init as initRouter } from './modules/router.js';

// Page-specific module loaders
const PAGE_MODULES = {
    search: () => import('./search-page.js'),
    about: () => import('./contact.js'),
    compose: () => import('./compose.js'),
    admin_home: () => import('./admin.js'),
};

let currentPageModule = null;

async function loadPageModule(template) {
    // Cleanup previous module if it supports it
    if (currentPageModule?.destroy) currentPageModule.destroy();
    currentPageModule = null;

    const loader = PAGE_MODULES[template];
    if (loader) {
        const mod = await loader();
        mod.init();
        currentPageModule = mod;
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

document.addEventListener('DOMContentLoaded', () => {
    // Shell modules — run once, persist across navigations
    initNavigation();
    initTheme();
    initScroll();
    initLogin();
    initToast();
    initFab();
    initComposeSheet();

    // Content modules — initial page
    initHighlight();
    initLazy();

    // Page-specific module — initial page
    loadPageModule(document.body.dataset.template);

    // Router — last, passes reinitPage as callback
    initRouter(reinitPage);
});
