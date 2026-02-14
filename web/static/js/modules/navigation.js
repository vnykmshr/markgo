/**
 * Navigation — mobile menu toggle, bottom nav, active link highlighting, scroll shadow.
 */

function updateActiveLinks() {
    const currentPath = window.location.pathname;

    // Top nav + footer link active states
    document.querySelectorAll('.nav-link, .footer-link').forEach((link) => {
        try {
            const linkPath = new URL(link.href, location.origin).pathname;
            link.classList.toggle('active',
                linkPath === currentPath ||
                (currentPath.startsWith('/writing/') && linkPath === '/writing') ||
                (currentPath.startsWith('/tags/') && linkPath === '/tags') ||
                (currentPath.startsWith('/categories/') && linkPath === '/categories')
            );
        } catch (err) { console.debug('Skipping link active state:', link.href, err.message); }
    });

    // Bottom nav active states
    document.querySelectorAll('.bottom-nav-item[data-nav]').forEach((item) => {
        const nav = item.dataset.nav;
        let isActive = false;
        if (nav === 'home') isActive = currentPath === '/';
        else if (nav === 'writing') isActive = currentPath === '/writing' || currentPath.startsWith('/writing/');
        else if (nav === 'search') isActive = currentPath === '/search';
        else if (nav === 'about') isActive = currentPath === '/about';
        item.classList.toggle('active', isActive);
    });
}

export function init() {
    const navbar = document.querySelector('.navbar');

    // Bottom nav compose button → dispatch fab:compose
    const composeBtn = document.querySelector('.bottom-nav-compose');
    if (composeBtn) {
        composeBtn.addEventListener('click', () => {
            document.dispatchEvent(new CustomEvent('fab:compose'));
        });
    }

    // Bottom nav AMA button → dispatch fab:ama
    const amaBtn = document.querySelector('.bottom-nav-ama');
    if (amaBtn) {
        amaBtn.addEventListener('click', () => {
            document.dispatchEvent(new CustomEvent('fab:ama'));
        });
    }

    // Active link highlighting
    updateActiveLinks();

    // Re-highlight after SPA navigation
    document.addEventListener('router:navigate-end', updateActiveLinks);

    // Navbar scroll shadow
    let scrollTicking = false;
    window.addEventListener('scroll', () => {
        if (!scrollTicking) {
            requestAnimationFrame(() => {
                if (navbar) {
                    navbar.classList.toggle('scrolled', window.scrollY > 50);
                }
                scrollTicking = false;
            });
            scrollTicking = true;
        }
    }, { passive: true });
}
