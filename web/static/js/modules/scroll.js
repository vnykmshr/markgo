/**
 * Scroll behavior — back-to-top button + smooth anchor scrolling.
 */

export function init() {
    // Create back-to-top button
    const btn = document.createElement('button');
    btn.className = 'back-to-top';
    btn.textContent = '\u2191';
    btn.title = 'Back to top';
    btn.style.display = 'none';
    btn.setAttribute('aria-label', 'Back to top');
    document.body.appendChild(btn);

    // Show/hide on scroll
    window.addEventListener('scroll', () => {
        btn.style.display = window.scrollY > 500 ? 'block' : 'none';
    }, { passive: true });

    // Smooth scroll to top
    btn.addEventListener('click', () => {
        window.scrollTo({ top: 0, behavior: 'smooth' });
    });

    // Smooth scroll for anchor links
    document.addEventListener('click', (e) => {
        if (e.target.matches('a[href^="#"]')) {
            e.preventDefault();
            const target = document.querySelector(e.target.getAttribute('href'));
            if (target) {
                target.scrollIntoView({ behavior: 'smooth', block: 'start' });
            }
        }
    });
}
