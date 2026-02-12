/**
 * Navigation — mobile menu toggle, active link highlighting, scroll shadow.
 */

export function init() {
    const navbar = document.querySelector('.navbar');
    const navbarToggle = document.getElementById('navbar-toggle');
    const navbarMenu = document.getElementById('navbar-menu');

    // Mobile menu toggle
    if (navbarToggle && navbarMenu) {
        navbarToggle.addEventListener('click', () => {
            navbarMenu.classList.toggle('active');
            navbarToggle.classList.toggle('active');
            navbarToggle.setAttribute('aria-expanded', navbarMenu.classList.contains('active'));
        });

        // Close menu when clicking outside
        document.addEventListener('click', (event) => {
            if (!navbar.contains(event.target)) {
                navbarMenu.classList.remove('active');
                navbarToggle.classList.remove('active');
                navbarToggle.setAttribute('aria-expanded', 'false');
            }
        });

        // Close menu on Escape
        document.addEventListener('keydown', (event) => {
            if (event.key === 'Escape') {
                navbarMenu.classList.remove('active');
                navbarToggle.classList.remove('active');
                navbarToggle.setAttribute('aria-expanded', 'false');
            }
        });
    }

    // Active navigation link highlighting
    const currentPath = window.location.pathname;
    document.querySelectorAll('.nav-link').forEach((link) => {
        const linkPath = new URL(link.href).pathname;
        if (
            linkPath === currentPath ||
            (currentPath.startsWith('/writing/') && linkPath === '/writing') ||
            (currentPath.startsWith('/tags/') && linkPath === '/tags') ||
            (currentPath.startsWith('/categories/') && linkPath === '/categories')
        ) {
            link.classList.add('active');
        }
    });

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
