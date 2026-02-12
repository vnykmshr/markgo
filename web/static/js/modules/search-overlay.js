/**
 * Search Overlay — full-screen search on mobile.
 *
 * Shell module: persists across SPA navigations.
 * Opens from bottom nav search tap or Cmd/Ctrl+K shortcut.
 * Submits via SPA router navigation to /search?q=...
 */

let overlay = null;
let input = null;
let isOpen = false;
let viewportHandler = null;

function buildOverlay() {
    const el = document.createElement('div');
    el.className = 'search-overlay';
    el.setAttribute('role', 'dialog');
    el.setAttribute('aria-modal', 'true');
    el.setAttribute('aria-label', 'Search');
    el.hidden = true;

    const backdrop = document.createElement('div');
    backdrop.className = 'search-overlay-backdrop';
    backdrop.addEventListener('click', close);

    const panel = document.createElement('div');
    panel.className = 'search-overlay-panel';

    const header = document.createElement('div');
    header.className = 'search-overlay-header';

    const form = document.createElement('form');
    form.className = 'search-overlay-form';
    form.addEventListener('submit', handleSubmit);

    input = document.createElement('input');
    input.type = 'search';
    input.className = 'search-overlay-input';
    input.placeholder = 'Search posts\u2026';
    input.autocomplete = 'off';

    const cancelBtn = document.createElement('button');
    cancelBtn.type = 'button';
    cancelBtn.className = 'search-overlay-cancel';
    cancelBtn.textContent = 'Cancel';
    cancelBtn.addEventListener('click', close);

    form.appendChild(input);
    header.appendChild(form);
    header.appendChild(cancelBtn);
    panel.appendChild(header);

    el.appendChild(backdrop);
    el.appendChild(panel);

    return el;
}

function handleSubmit(e) {
    e.preventDefault();
    const query = input.value.trim();
    if (!query) return;

    close();
    // Navigate via SPA router — synthetic link click triggers router interception
    const a = document.createElement('a');
    a.href = '/search?q=' + encodeURIComponent(query);
    a.style.display = 'none';
    document.body.appendChild(a);
    a.click();
    a.remove();
}

function open() {
    if (isOpen) return;
    if (!overlay) {
        overlay = buildOverlay();
        document.body.appendChild(overlay);
    }

    overlay.hidden = false;
    isOpen = true;
    document.body.style.overflow = 'hidden';

    requestAnimationFrame(() => input.focus());
    document.addEventListener('keydown', handleKeydown);

    // Visual viewport — reposition above virtual keyboard on iOS
    if (window.visualViewport) {
        viewportHandler = () => {
            if (!overlay || !overlay.isConnected) return;
            const vv = window.visualViewport;
            overlay.style.height = vv.height + 'px';
            overlay.style.top = vv.offsetTop + 'px';
        };
        viewportHandler();
        window.visualViewport.addEventListener('resize', viewportHandler);
        window.visualViewport.addEventListener('scroll', viewportHandler);
    }
}

function close() {
    if (!isOpen) return;
    isOpen = false;
    overlay.hidden = true;
    overlay.style.height = '';
    overlay.style.top = '';
    document.body.style.overflow = '';
    input.value = '';
    document.removeEventListener('keydown', handleKeydown);

    if (viewportHandler && window.visualViewport) {
        window.visualViewport.removeEventListener('resize', viewportHandler);
        window.visualViewport.removeEventListener('scroll', viewportHandler);
        viewportHandler = null;
    }
}

function handleKeydown(e) {
    if (e.key === 'Escape') {
        e.preventDefault();
        close();
    }
}

export function init() {
    // Bottom nav search → open overlay on mobile
    const searchNavItem = document.querySelector('.bottom-nav-item[data-nav="search"]');
    if (searchNavItem) {
        searchNavItem.addEventListener('click', (e) => {
            // Only intercept on mobile (when bottom nav is visible)
            if (window.innerWidth < 769) {
                e.preventDefault();
                open();
            }
        });
    }

    // Cmd/Ctrl+K shortcut
    document.addEventListener('keydown', (e) => {
        if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
            e.preventDefault();
            if (isOpen) close(); else open();
        }
    });

    // Close on SPA navigation
    document.addEventListener('router:navigate-start', close);
}
