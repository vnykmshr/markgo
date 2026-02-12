/**
 * Lazy loading — IntersectionObserver for images with data-src.
 */

export function init() {
    try {
        if (!('IntersectionObserver' in window)) {
            // Fallback: load all images immediately
            document.querySelectorAll('img[data-src]').forEach((img) => {
                img.src = img.dataset.src;
                img.classList.remove('lazy');
            });
            return;
        }

        const observer = new IntersectionObserver((entries) => {
            entries.forEach((entry) => {
                if (!entry.isIntersecting) return;
                const img = entry.target;

                img.onerror = () => {
                    img.classList.add('lazy-error');
                };
                img.onload = () => {
                    img.classList.remove('lazy');
                    img.classList.add('lazy-loaded');
                };

                img.src = img.dataset.src;
                observer.unobserve(img);
            });
        });

        document.querySelectorAll('img[data-src]').forEach((img) => {
            img.classList.add('lazy');
            observer.observe(img);
        });
    } catch (error) {
        console.error('Lazy loading initialization failed:', error);
        // Fallback: load all images immediately
        document.querySelectorAll('img[data-src]').forEach((img) => {
            img.src = img.dataset.src;
        });
    }
}
