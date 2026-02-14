/**
 * Focus trap utility for modal dialogs.
 *
 * Traps Tab/Shift+Tab inside a container so keyboard users
 * cannot escape into background content while a dialog is open.
 *
 * Usage:
 *   import { trapFocus, releaseFocus } from './focus-trap.js';
 *   trapFocus(dialogElement);   // on open
 *   releaseFocus();             // on close
 */

const FOCUSABLE = 'a[href], button:not([disabled]), textarea, input:not([type="hidden"]):not([tabindex="-1"]), select, [tabindex]:not([tabindex="-1"])';

let activeContainer = null;
let handler = null;

function onKeydown(e) {
    if (e.key !== 'Tab' || !activeContainer) return;

    const focusable = [...activeContainer.querySelectorAll(FOCUSABLE)];
    if (focusable.length === 0) return;

    const first = focusable[0];
    const last = focusable[focusable.length - 1];

    if (e.shiftKey) {
        if (document.activeElement === first) {
            last.focus();
            e.preventDefault();
        }
    } else {
        if (document.activeElement === last) {
            first.focus();
            e.preventDefault();
        }
    }
}

/**
 * Trap focus inside a container element.
 * Only one trap can be active at a time.
 */
export function trapFocus(container) {
    releaseFocus();
    activeContainer = container;
    handler = onKeydown;
    document.addEventListener('keydown', handler);
}

/**
 * Release the active focus trap.
 */
export function releaseFocus() {
    if (handler) {
        document.removeEventListener('keydown', handler);
        handler = null;
    }
    activeContainer = null;
}
