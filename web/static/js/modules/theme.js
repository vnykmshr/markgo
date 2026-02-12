/**
 * Theme toggle — light/dark mode with localStorage persistence.
 *
 * Two independent axes:
 * 1. Color theme: data-color-theme attribute (server-rendered from BLOG_THEME config)
 * 2. Light/dark mode: data-theme attribute ("dark", "light", or absent for auto)
 */

function createSvgPath(d) {
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('width', '16');
    svg.setAttribute('height', '16');
    svg.setAttribute('fill', 'currentColor');
    svg.setAttribute('viewBox', '0 0 16 16');
    const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    path.setAttribute('d', d);
    svg.appendChild(path);
    return svg;
}

// Sun icon paths (multiple paths needed)
function createSunIcon() {
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('width', '16');
    svg.setAttribute('height', '16');
    svg.setAttribute('fill', 'currentColor');
    svg.setAttribute('viewBox', '0 0 16 16');
    const paths = [
        'M8 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM8 0a.5.5 0 0 1 .5.5v2a.5.5 0 0 1-1 0v-2A.5.5 0 0 1 8 0zm0 13a.5.5 0 0 1 .5.5v2a.5.5 0 0 1-1 0v-2A.5.5 0 0 1 8 13zm8-5a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1 0-1h2a.5.5 0 0 1 .5.5zM3 8a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1 0-1h2A.5.5 0 0 1 3 8zm10.657-5.657a.5.5 0 0 1 0 .707l-1.414 1.415a.5.5 0 1 1-.707-.708l1.414-1.414a.5.5 0 0 1 .707 0zm-9.193 9.193a.5.5 0 0 1 0 .707L3.05 13.657a.5.5 0 0 1-.707-.707l1.414-1.414a.5.5 0 0 1 .707 0zm9.193 2.121a.5.5 0 0 1-.707 0l-1.414-1.414a.5.5 0 0 1 .707-.707l1.414 1.414a.5.5 0 0 1 0 .707zM4.464 4.465a.5.5 0 0 1-.707 0L2.343 3.05a.5.5 0 1 1 .707-.707l1.414 1.414a.5.5 0 0 1 0 .708z',
    ];
    paths.forEach((d) => {
        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.setAttribute('d', d);
        svg.appendChild(path);
    });
    return svg;
}

const MOON_PATH = 'M6 .278a.768.768 0 0 1 .08.858 7.208 7.208 0 0 0-.878 3.46c0 4.021 3.278 7.277 7.318 7.277.527 0 1.04-.055 1.533-.16a.787.787 0 0 1 .81.316.733.733 0 0 1-.031.893A8.349 8.349 0 0 1 8.344 16C3.734 16 0 12.286 0 7.71 0 4.266 2.114 1.312 5.124.06A.752.752 0 0 1 6 .278z';

function resolveMode(mode) {
    if (mode === 'dark' || mode === 'light') return mode;
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
        return 'dark';
    }
    return 'light';
}

function applyTheme(mode) {
    if (mode === 'dark' || mode === 'light') {
        document.documentElement.setAttribute('data-theme', mode);
    } else {
        document.documentElement.removeAttribute('data-theme');
    }
    updateThemeColor();
}

function updateThemeColor() {
    const meta = document.querySelector('meta[name="theme-color"]');
    if (!meta) return;
    // Read computed background from body — respects light/dark/system preference
    const bg = getComputedStyle(document.body).backgroundColor;
    if (bg) meta.content = bg;
}

function getCurrentMode() {
    try {
        const saved = localStorage.getItem('theme');
        if (saved === 'dark' || saved === 'light') return saved;
    } catch (e) {
        // ignore
    }
    return resolveMode(null);
}

function updateToggleIcon(themeToggle, mode) {
    const oldIcon = themeToggle.querySelector('svg');
    if (!oldIcon) return;

    const newIcon = mode === 'dark' ? createSunIcon() : createSvgPath(MOON_PATH);
    themeToggle.replaceChild(newIcon, oldIcon);
    themeToggle.setAttribute('aria-label', mode === 'dark' ? 'Switch to light mode' : 'Switch to dark mode');
}

export function init() {
    const themeToggle = document.querySelector('.theme-toggle');
    if (!themeToggle) return;

    try {
        let savedMode = null;
        try { savedMode = localStorage.getItem('theme'); } catch (e) { /* ignore */ }

        applyTheme(savedMode);
        updateToggleIcon(themeToggle, resolveMode(savedMode));
        // Defer theme-color update to ensure computed styles are available
        requestAnimationFrame(updateThemeColor);

        themeToggle.addEventListener('click', () => {
            try {
                const newMode = getCurrentMode() === 'dark' ? 'light' : 'dark';
                try { localStorage.setItem('theme', newMode); } catch (e) { /* ignore */ }
                applyTheme(newMode);
                updateToggleIcon(themeToggle, newMode);
            } catch (err) {
                console.error('Theme toggle failed:', err);
            }
        });

        // Respond to system preference changes when no explicit choice
        if (window.matchMedia) {
            window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
                try {
                    const saved = localStorage.getItem('theme');
                    if (!saved) {
                        applyTheme(null);
                        updateToggleIcon(themeToggle, getCurrentMode());
                    }
                } catch (e) {
                    // ignore
                }
                requestAnimationFrame(updateThemeColor);
            });
        }
    } catch (error) {
        console.error('Theme toggle initialization failed:', error);
    }
}
