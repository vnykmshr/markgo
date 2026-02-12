/**
 * Theme popover — light/dark/auto mode + color presets.
 *
 * Two independent axes:
 * 1. Color theme: data-color-theme attribute (server-rendered from BLOG_THEME config)
 * 2. Light/dark mode: data-theme attribute ("dark", "light", or absent for auto)
 */

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
    const bg = getComputedStyle(document.body).backgroundColor;
    if (bg) meta.content = bg;
}

function getSavedMode() {
    try {
        const saved = localStorage.getItem('theme');
        if (saved === 'dark' || saved === 'light') return saved;
    } catch (e) {
        // ignore
    }
    return null;
}

// ---------------------------------------------------------------------------
// Color theme
// ---------------------------------------------------------------------------

function setColorTheme(preset) {
    const html = document.documentElement;
    if (preset && preset !== 'default') {
        html.setAttribute('data-color-theme', preset);
    } else {
        html.removeAttribute('data-color-theme');
    }
    try { localStorage.setItem('colorTheme', preset || 'default'); } catch (e) { /* ignore */ }
    requestAnimationFrame(updateThemeColor);
}

function updateSwatchActive(container, preset) {
    container.querySelectorAll('.color-swatch').forEach((btn) => {
        const isActive = btn.dataset.color === (preset || 'default');
        btn.classList.toggle('active', isActive);
        btn.setAttribute('aria-checked', isActive);
    });
}

// ---------------------------------------------------------------------------
// Mode buttons
// ---------------------------------------------------------------------------

function updateModeActive(container, mode) {
    container.querySelectorAll('.theme-mode-btn').forEach((btn) => {
        const isActive = btn.dataset.mode === (mode || 'auto');
        btn.classList.toggle('active', isActive);
        btn.setAttribute('aria-checked', isActive);
    });
}

// ---------------------------------------------------------------------------
// Init
// ---------------------------------------------------------------------------

export function init() {
    const trigger = document.querySelector('.theme-btn');
    const popover = document.getElementById('theme-popover');
    if (!trigger || !popover) return;

    try {
        const savedMode = getSavedMode();
        applyTheme(savedMode);
        requestAnimationFrame(updateThemeColor);

        // Restore saved color theme
        let savedColor = 'default';
        try { savedColor = localStorage.getItem('colorTheme') || 'default'; } catch (e) { /* ignore */ }
        setColorTheme(savedColor);

        // Set initial active states
        updateModeActive(popover, savedMode || 'auto');
        updateSwatchActive(popover, savedColor);

        // Popover toggle
        function openPopover() {
            popover.hidden = false;
            trigger.setAttribute('aria-expanded', 'true');
            const firstBtn = popover.querySelector('.theme-mode-btn');
            if (firstBtn) firstBtn.focus();
        }

        function closePopover() {
            popover.hidden = true;
            trigger.setAttribute('aria-expanded', 'false');
        }

        trigger.addEventListener('click', (e) => {
            e.stopPropagation();
            if (popover.hidden) {
                openPopover();
            } else {
                closePopover();
            }
        });

        // Close on click outside
        document.addEventListener('click', (e) => {
            if (!popover.hidden && !popover.contains(e.target) && e.target !== trigger) {
                closePopover();
            }
        });

        // Close on Escape
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && !popover.hidden) {
                closePopover();
                trigger.focus();
            }
        });

        // Mode selection
        popover.addEventListener('click', (e) => {
            const modeBtn = e.target.closest('.theme-mode-btn');
            if (modeBtn) {
                const mode = modeBtn.dataset.mode;
                if (mode === 'auto') {
                    try { localStorage.removeItem('theme'); } catch (err) { /* ignore */ }
                    applyTheme(null);
                } else {
                    try { localStorage.setItem('theme', mode); } catch (err) { /* ignore */ }
                    applyTheme(mode);
                }
                updateModeActive(popover, mode);
                return;
            }

            // Color selection
            const swatch = e.target.closest('.color-swatch');
            if (swatch) {
                const preset = swatch.dataset.color;
                setColorTheme(preset);
                updateSwatchActive(popover, preset);
            }
        });

        // Respond to system preference changes when in auto mode
        if (window.matchMedia) {
            window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
                const saved = getSavedMode();
                if (!saved) {
                    applyTheme(null);
                }
                requestAnimationFrame(updateThemeColor);
            });
        }
    } catch (error) {
        console.error('Theme initialization failed:', error);
    }
}
