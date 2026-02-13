/**
 * Popover — shared toggle logic for header popovers.
 *
 * Handles toggle, Escape, click-outside, link-click-close,
 * and popover:exclusive mutual exclusion.
 *
 * Returns a control object so callers can programmatically open/close.
 */

export function initPopover(popoverId, triggerSelector, onOpen, onClose) {
    const popover = document.getElementById(popoverId);
    const trigger = document.querySelector(triggerSelector);
    if (!popover || !trigger) return null;

    const ac = new AbortController();
    const { signal } = ac;

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
    }, { signal });

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && !popover.hidden) {
            close();
            trigger.focus();
        }
    }, { signal });

    document.addEventListener('click', (e) => {
        if (!popover.hidden && !popover.contains(e.target) && e.target !== trigger) close();
    }, { signal });

    // Close on link click (SPA navigation) and let event bubble to router
    popover.addEventListener('click', (e) => {
        if (e.target.closest('a[href]')) close();
    }, { signal });

    // Mutual exclusion — close when another popover opens
    document.addEventListener('popover:exclusive', (e) => {
        if (e.detail !== popoverId && !popover.hidden) close();
    }, { signal });

    return { open, close, destroy: () => ac.abort(), popover, trigger };
}
