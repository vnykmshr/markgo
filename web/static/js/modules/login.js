/**
 * Login — handles login popover (public), auth gate (protected), and admin popover (authenticated).
 * Login forms share .login-form class and POST to /login via fetch.
 * Admin popover provides quick access to dashboard, drafts, sign out.
 *
 * Reactive auth: on successful login from the popover, swaps login trigger/popover
 * to admin trigger/popover in-place (no full-page reload). Dispatches auth:statechange
 * so other modules can react (e.g. draft recovery toast).
 */

import { initPopover } from './popover.js';
import { showToast } from './toast.js';

const DRAFT_KEY = 'markgo:compose-draft';

export function init() {
    // Attach fetch-based submit to all login forms (popover + auth gate)
    document.querySelectorAll('.login-form').forEach((form) => {
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            const errorEl = form.querySelector('.login-error');
            if (errorEl) {
                errorEl.hidden = true;
                errorEl.textContent = '';
            }

            const submitBtn = form.querySelector('button[type="submit"]');
            if (submitBtn) submitBtn.disabled = true;

            const isPopover = form.closest('.login-popover') !== null;

            fetch('/login', {
                method: 'POST',
                headers: { Accept: 'application/json' },
                body: new FormData(form),
                credentials: 'same-origin',
            })
                .then((res) => {
                    if (res.status === 403) throw new Error('Session expired. Please refresh the page and try again.');
                    if (res.status === 429) throw new Error('Too many attempts. Please wait and try again.');
                    return res.json().then(
                        (data) => ({ ok: res.ok, data }),
                        () => { throw new Error('Server error. Please refresh the page.'); }
                    );
                })
                .then((result) => {
                    if (result.data.success) {
                        if (isPopover) {
                            // Reactive auth — swap UI in place
                            const swapped = swapToAuthenticatedUI();
                            if (swapped !== false) {
                                document.dispatchEvent(new CustomEvent('auth:statechange', { detail: { authenticated: true } }));
                                checkDraftRecovery();
                            }
                        } else {
                            // Auth gate — full reload to render protected page
                            window.location.href = result.data.redirect || window.location.pathname;
                        }
                    } else {
                        if (errorEl) {
                            errorEl.textContent = result.data.error || 'Login failed.';
                            errorEl.hidden = false;
                        }
                        if (submitBtn) submitBtn.disabled = false;
                    }
                })
                .catch((err) => {
                    if (errorEl) {
                        errorEl.textContent = err.message || 'Network error. Please try again.';
                        errorEl.hidden = false;
                    }
                    if (submitBtn) submitBtn.disabled = false;
                });
        });
    });

    // Auto-focus the inline auth gate form (protected pages)
    const authGateInput = document.querySelector('.auth-gate-form input[name="username"]');
    if (authGateInput) authGateInput.focus();

    // Login popover toggle (unauthenticated, public pages)
    initPopover('login-popover', '.login-trigger', (popover) => {
        const firstInput = popover.querySelector('input[name="username"]');
        if (firstInput) firstInput.focus();
    }, (popover) => {
        const errorEl = popover.querySelector('.login-error');
        if (errorEl) { errorEl.hidden = true; errorEl.textContent = ''; }
    });

    // Admin popover toggle (authenticated)
    initPopover('admin-popover', '.admin-trigger');

    // Listen for programmatic open-login requests (e.g. from compose 401)
    document.addEventListener('auth:open-login', () => {
        const loginTrigger = document.querySelector('.login-trigger');
        if (loginTrigger) {
            loginTrigger.click();
        } else {
            // Login trigger not in DOM (e.g. after reactive swap to admin)
            window.location.href = '/login';
        }
    });
}

/**
 * Create the person SVG icon used by both login and admin triggers.
 * Built with DOM API to avoid innerHTML.
 */
function createPersonIcon() {
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('width', '16');
    svg.setAttribute('height', '16');
    svg.setAttribute('fill', 'currentColor');
    svg.setAttribute('viewBox', '0 0 16 16');
    const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    path.setAttribute('d', 'M8 8a3 3 0 1 0 0-6 3 3 0 0 0 0 6zm2-3a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm4 8c0 1-1 1-1 1H3s-1 0-1-1 1-4 6-4 6 3 6 4zm-1-.004c-.001-.246-.154-.986-.832-1.664C11.516 10.68 10.289 10 8 10c-2.29 0-3.516.68-4.168 1.332-.678.678-.83 1.418-.832 1.664h10z');
    svg.appendChild(path);
    return svg;
}

/**
 * Swap login trigger + popover to admin trigger + popover.
 * Old DOM nodes and their event listeners are garbage-collected.
 */
function swapToAuthenticatedUI() {
    const loginTrigger = document.querySelector('.login-trigger');
    const loginPopover = document.getElementById('login-popover');
    if (!loginTrigger) {
        // DOM is in unexpected state — fall back to full reload
        console.warn('login-trigger not found for reactive auth swap — reloading');
        window.location.reload();
        return false;
    }

    // Create admin trigger button
    const adminTrigger = document.createElement('button');
    adminTrigger.className = 'nav-action-btn admin-trigger';
    adminTrigger.setAttribute('aria-label', 'Account');
    adminTrigger.setAttribute('aria-expanded', 'false');
    adminTrigger.setAttribute('aria-haspopup', 'true');
    adminTrigger.title = 'Account';
    adminTrigger.appendChild(createPersonIcon());

    // Replace login trigger with admin trigger
    loginTrigger.replaceWith(adminTrigger);

    // Build admin popover with DOM API
    const adminPopover = document.createElement('div');
    adminPopover.className = 'admin-popover';
    adminPopover.id = 'admin-popover';
    adminPopover.hidden = true;

    const nav = document.createElement('nav');
    nav.className = 'admin-popover-nav';
    nav.setAttribute('aria-label', 'Admin');

    const dashLink = document.createElement('a');
    dashLink.href = '/admin';
    dashLink.className = 'admin-popover-link';
    dashLink.textContent = 'Dashboard';

    const writingLink = document.createElement('a');
    writingLink.href = '/writing';
    writingLink.className = 'admin-popover-link';
    writingLink.textContent = 'Writing';

    const draftsLink = document.createElement('a');
    draftsLink.href = '/admin/drafts';
    draftsLink.className = 'admin-popover-link';
    draftsLink.textContent = 'Drafts';

    nav.appendChild(dashLink);
    nav.appendChild(writingLink);
    nav.appendChild(draftsLink);

    const divider = document.createElement('div');
    divider.className = 'admin-popover-divider';

    const logoutLink = document.createElement('a');
    logoutLink.href = '/logout';
    logoutLink.className = 'admin-popover-link admin-popover-logout';
    logoutLink.textContent = 'Sign out';

    adminPopover.appendChild(nav);
    adminPopover.appendChild(divider);
    adminPopover.appendChild(logoutLink);

    // Replace login popover with admin popover
    if (loginPopover) {
        loginPopover.replaceWith(adminPopover);
    } else {
        console.warn('login-popover not found — using fallback container append');
        const container = adminTrigger.closest('.container');
        if (container) {
            container.appendChild(adminPopover);
        } else {
            console.warn('No .container found for admin popover — admin menu non-functional');
            showToast('Please refresh the page to access your account menu', 'warning');
            return false;
        }
    }

    // Initialize popover behavior on new elements
    initPopover('admin-popover', '.admin-trigger');
    return true;
}

/**
 * After login, check if there's a preserved quick compose draft
 * and surface it via toast.
 */
function checkDraftRecovery() {
    try {
        const raw = localStorage.getItem(DRAFT_KEY);
        if (!raw) return;
        const draft = JSON.parse(raw);
        if (draft && (draft.content || draft.title)) {
            showToast('Draft preserved \u2014 tap + to continue', 'info');
        }
    } catch (err) {
        console.warn('Draft recovery check failed:', err.message);
    }
}
