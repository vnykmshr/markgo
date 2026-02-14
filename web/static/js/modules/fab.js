/**
 * Floating Action Button — dual-purpose trigger.
 *
 * Shell module: runs once, persists across SPA navigations.
 * Auth → "+" compose button, dispatches "fab:compose".
 * Unauth → "?" AMA button, dispatches "fab:ama".
 */

let fabEl = null;

function isAuthenticated() {
    return document.body.dataset.authenticated === 'true';
}

function createComposeFAB() {
    const btn = document.createElement('button');
    btn.className = 'fab';
    btn.setAttribute('aria-label', 'New post');
    btn.title = 'New post';

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

function createAMAFAB() {
    const btn = document.createElement('button');
    btn.className = 'fab fab-ama';
    btn.setAttribute('aria-label', 'Ask a question');
    btn.title = 'Ask a question';

    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('width', '24');
    svg.setAttribute('height', '24');
    svg.setAttribute('viewBox', '0 0 16 16');
    svg.setAttribute('fill', 'currentColor');
    svg.setAttribute('aria-hidden', 'true');

    const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    path.setAttribute('d', 'M5.255 5.786a.237.237 0 0 0 .241.247h.825c.138 0 .248-.113.266-.25.09-.656.54-1.134 1.342-1.134.686 0 1.314.343 1.314 1.168 0 .635-.374.927-.965 1.371-.673.489-1.206 1.06-1.168 1.987l.003.217a.25.25 0 0 0 .25.246h.811a.25.25 0 0 0 .25-.25v-.105c0-.718.273-.927 1.01-1.486.609-.463 1.244-.977 1.244-2.056 0-1.511-1.276-2.241-2.673-2.241-1.267 0-2.655.59-2.75 2.286zm1.557 5.763c0 .533.425.927 1.01.927.609 0 1.028-.394 1.028-.927 0-.552-.42-.94-1.029-.94-.584 0-1.009.388-1.009.94z');

    svg.appendChild(path);
    btn.appendChild(svg);

    btn.addEventListener('click', () => {
        document.dispatchEvent(new CustomEvent('fab:ama'));
    });

    return btn;
}

function registerKeyboardShortcut() {
    document.addEventListener('keydown', (e) => {
        // Cmd/Ctrl+. — open compose (auth) or AMA (unauth)
        // Avoids Cmd+N which is browser "new window"
        if ((e.metaKey || e.ctrlKey) && e.key === '.') {
            e.preventDefault();
            if (isAuthenticated()) {
                document.dispatchEvent(new CustomEvent('fab:compose'));
            } else {
                document.dispatchEvent(new CustomEvent('fab:ama'));
            }
        }
    });
}

function upgradeFAB() {
    // Replace AMA FAB with compose FAB after login
    if (fabEl) {
        fabEl.remove();
    }
    fabEl = createComposeFAB();
    document.body.appendChild(fabEl);
}

export function init() {
    if (fabEl) return; // already initialized

    if (isAuthenticated()) {
        fabEl = createComposeFAB();
    } else {
        fabEl = createAMAFAB();
        // Upgrade to compose FAB if user logs in
        document.addEventListener('auth:authenticated', () => upgradeFAB(), { once: true });
    }

    document.body.appendChild(fabEl);
    registerKeyboardShortcut();
}
