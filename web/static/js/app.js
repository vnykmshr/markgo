/**
 * MarkGo Engine — ES Module Entry Point
 * Imports core modules and lazily loads page-specific modules
 * based on the data-template attribute on <body>.
 */

import { init as initNavigation } from './modules/navigation.js';
import { init as initTheme } from './modules/theme.js';
import { init as initHighlight } from './modules/highlight.js';
import { init as initScroll } from './modules/scroll.js';
import { init as initLazy } from './modules/lazy.js';
import { init as initLogin } from './modules/login.js';

document.addEventListener('DOMContentLoaded', () => {
    // Core modules — run on every page
    initNavigation();
    initTheme();
    initHighlight();
    initScroll();
    initLazy();
    initLogin();

    // Page-specific modules — lazily imported based on template
    const template = document.body.dataset.template;

    if (template === 'search') {
        import('./search-page.js').then((m) => m.init());
    }
    if (template === 'about') {
        import('./contact.js').then((m) => m.init());
    }
    if (template === 'compose') {
        import('./compose.js').then((m) => m.init());
    }
    if (template === 'admin_home') {
        import('./admin.js').then((m) => m.init());
    }
});
