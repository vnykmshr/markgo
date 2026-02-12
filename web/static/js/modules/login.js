/**
 * Login — handles login popover (public), auth gate (protected), and admin popover (authenticated).
 * Login forms share .login-form class and POST to /login via fetch.
 * Admin popover provides quick access to dashboard, compose, drafts, sign out.
 */

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
                        window.location.href = result.data.redirect || window.location.pathname;
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
}

function initPopover(popoverId, triggerSelector, onOpen, onClose) {
    const popover = document.getElementById(popoverId);
    const trigger = document.querySelector(triggerSelector);
    if (!popover || !trigger) return;

    function open() {
        // Close other popovers before opening this one
        document.dispatchEvent(new CustomEvent('popover:exclusive', { detail: popoverId }));
        popover.hidden = false;
        trigger.setAttribute('aria-expanded', 'true');
        if (onOpen) onOpen(popover);
    }

    function close() {
        popover.hidden = true;
        trigger.setAttribute('aria-expanded', 'false');
        if (onClose) onClose(popover);
    }

    trigger.addEventListener('click', (e) => {
        e.stopPropagation();
        if (popover.hidden) open(); else close();
    });

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && !popover.hidden) {
            close();
            trigger.focus();
        }
    });

    document.addEventListener('click', (e) => {
        if (!popover.hidden && !popover.contains(e.target) && e.target !== trigger) close();
    });

    // Close on link click (SPA navigation) and let event bubble to router
    popover.addEventListener('click', (e) => {
        if (e.target.closest('a[href]')) close();
    });

    // Mutual exclusion — close when another popover opens
    document.addEventListener('popover:exclusive', (e) => {
        if (e.detail !== popoverId && !popover.hidden) close();
    });
}
