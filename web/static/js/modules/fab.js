/**
 * Floating Action Button — auth-gated compose trigger.
 *
 * Shell module: runs once, persists across SPA navigations.
 * Shows a "+" button in the bottom-right for authenticated users.
 * Dispatches "fab:compose" custom event on click for the compose sheet.
 */

let fabEl = null;

function isAuthenticated() {
    return document.querySelector('a[href="/logout"]') !== null;
}

function createFAB() {
    const btn = document.createElement('button');
    btn.className = 'fab';
    btn.setAttribute('aria-label', 'New post');
    btn.title = 'New post';

    // Plus icon via SVG
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('width', '24');
    svg.setAttribute('height', '24');
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('fill', 'none');
    svg.setAttribute('stroke', 'currentColor');
    svg.setAttribute('stroke-width', '2.5');
    svg.setAttribute('stroke-linecap', 'round');
    svg.setAttribute('aria-hidden', 'true');

    const line1 = document.createElementNS('http://www.w3.org/2000/svg', 'line');
    line1.setAttribute('x1', '12');
    line1.setAttribute('y1', '5');
    line1.setAttribute('x2', '12');
    line1.setAttribute('y2', '19');

    const line2 = document.createElementNS('http://www.w3.org/2000/svg', 'line');
    line2.setAttribute('x1', '5');
    line2.setAttribute('y1', '12');
    line2.setAttribute('x2', '19');
    line2.setAttribute('y2', '12');

    svg.appendChild(line1);
    svg.appendChild(line2);
    btn.appendChild(svg);

    btn.addEventListener('click', () => {
        document.dispatchEvent(new CustomEvent('fab:compose'));
    });

    return btn;
}

export function init() {
    if (!isAuthenticated()) return;
    if (fabEl) return; // already initialized

    fabEl = createFAB();
    document.body.appendChild(fabEl);

    // Keyboard shortcut: Cmd/Ctrl+N opens compose sheet
    document.addEventListener('keydown', (e) => {
        if ((e.metaKey || e.ctrlKey) && e.key === 'n') {
            e.preventDefault();
            document.dispatchEvent(new CustomEvent('fab:compose'));
        }
    });
}
