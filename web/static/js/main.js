/**
 * MarkGo Engine Main JavaScript
 * Modern, vanilla JavaScript functionality for the Go blog engine
 */

(function () {
  "use strict";

  // DOM Content Loaded
  document.addEventListener("DOMContentLoaded", function () {
    initializeApp();
  });

  /**
   * Initialize the application
   */
  function initializeApp() {
    initNavigation();
    initSyntaxHighlighting();
    initSearchFunctionality();
    initThemeToggle();
    initLoginPopover();
    initScrollBehavior();
    initLazyLoading();
  }

  /**
   * Navigation functionality
   */
  function initNavigation() {
    const navbar = document.querySelector(".navbar");
    const navbarToggle = document.getElementById("navbar-toggle");
    const navbarMenu = document.getElementById("navbar-menu");

    // Mobile menu toggle
    if (navbarToggle && navbarMenu) {
      navbarToggle.addEventListener("click", function () {
        navbarMenu.classList.toggle("active");
        navbarToggle.classList.toggle("active");

        // Update ARIA attributes
        const isExpanded = navbarMenu.classList.contains("active");
        navbarToggle.setAttribute("aria-expanded", isExpanded);
      });

      // Close menu when clicking outside
      document.addEventListener("click", function (event) {
        if (!navbar.contains(event.target)) {
          navbarMenu.classList.remove("active");
          navbarToggle.classList.remove("active");
          navbarToggle.setAttribute("aria-expanded", "false");
        }
      });

      // Close menu on escape key
      document.addEventListener("keydown", function (event) {
        if (event.key === "Escape") {
          navbarMenu.classList.remove("active");
          navbarToggle.classList.remove("active");
          navbarToggle.setAttribute("aria-expanded", "false");
        }
      });
    }

    // Active navigation link highlighting
    const navLinks = document.querySelectorAll(".nav-link");
    const currentPath = window.location.pathname;

    navLinks.forEach((link) => {
      const linkPath = new URL(link.href).pathname;
      if (
        linkPath === currentPath ||
        (currentPath.startsWith("/writing/") && linkPath === "/writing") ||
        (currentPath.startsWith("/tags/") && linkPath === "/tags") ||
        (currentPath.startsWith("/categories/") && linkPath === "/categories")
      ) {
        link.classList.add("active");
      }
    });

    // Simple navbar scroll behavior - just add shadow
    var scrollTicking = false;
    window.addEventListener("scroll", function () {
      if (!scrollTicking) {
        requestAnimationFrame(function () {
          if (navbar) {
            if (window.scrollY > 50) {
              navbar.classList.add("scrolled");
            } else {
              navbar.classList.remove("scrolled");
            }
          }
          scrollTicking = false;
        });
        scrollTicking = true;
      }
    }, { passive: true });
  }

  /**
   * Syntax highlighting initialization
   */
  function initSyntaxHighlighting() {
    if (typeof hljs !== "undefined") {
      // Configure highlight.js
      hljs.configure({
        languages: [
          "javascript",
          "typescript",
          "go",
          "html",
          "css",
          "bash",
          "json",
          "yaml",
          "markdown",
        ],
      });

      // Highlight all code blocks
      hljs.highlightAll();

      // Add copy buttons to code blocks
      const codeBlocks = document.querySelectorAll("pre code");
      codeBlocks.forEach(function (codeBlock) {
        addCopyButton(codeBlock.parentElement);
      });
    }
  }

  /**
   * Add copy button to code block
   */
  function addCopyButton(preElement) {
    const button = document.createElement("button");
    button.className = "copy-code-btn";
    button.innerHTML = `
            <svg width="16" height="16" fill="currentColor" viewBox="0 0 16 16">
                <path d="M4 1.5H3a2 2 0 0 0-2 2V14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V3.5a2 2 0 0 0-2-2h-1v1h1a1 1 0 0 1 1 1V14a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V3.5a1 1 0 0 1 1-1h1v-1z"/>
                <path d="M9.5 1a.5.5 0 0 1 .5.5v1a.5.5 0 0 1-.5.5h-3a.5.5 0 0 1-.5-.5v-1a.5.5 0 0 1 .5-.5h3zm-3-1A1.5 1.5 0 0 0 5 1.5v1A1.5 1.5 0 0 0 6.5 4h3A1.5 1.5 0 0 0 11 2.5v-1A1.5 1.5 0 0 0 9.5 0h-3z"/>
            </svg>
        `;

    button.title = "Copy code to clipboard";

    button.addEventListener("click", function () {
      const code = preElement.querySelector("code").textContent;
      copyToClipboard(code)
        .then(function () {
          button.innerHTML = `
              <svg width="16" height="16" fill="currentColor" viewBox="0 0 16 16">
                  <path d="M10.97 4.97a.75.75 0 0 1 1.07 1.05l-3.99 4.99a.75.75 0 0 1-1.08.02L4.324 8.384a.75.75 0 1 1 1.06-1.06l2.094 2.093 3.473-4.425a.267.267 0 0 1 .02-.022z"/>
              </svg>
          `;
          button.classList.add("success");

          setTimeout(function () {
            button.innerHTML = `
                        <svg width="16" height="16" fill="currentColor" viewBox="0 0 16 16">
                            <path d="M4 1.5H3a2 2 0 0 0-2 2V14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V3.5a2 2 0 0 0-2-2h-1v1h1a1 1 0 0 1 1 1V14a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V3.5a1 1 0 0 1 1-1h1v-1z"/>
                            <path d="M9.5 1a.5.5 0 0 1 .5.5v1a.5.5 0 0 1-.5.5h-3a.5.5 0 0 1-.5-.5v-1a.5.5 0 0 1 .5-.5h3zm-3-1A1.5 1.5 0 0 0 5 1.5v1A1.5 1.5 0 0 0 6.5 4h3A1.5 1.5 0 0 0 11 2.5v-1A1.5 1.5 0 0 0 9.5 0h-3z"/>
                        </svg>
                    `;
            button.classList.remove("success");
          }, 2000);
        })
        .catch(function (err) {
          console.error("Failed to copy code:", err);
        });
    });

    preElement.style.position = "relative";
    preElement.appendChild(button);
  }

  /**
   * Search functionality
   */
  function initSearchFunctionality() {
    const searchForm = document.querySelector(".search-form");
    const searchInput = document.querySelector(".search-input");
    const searchResults = document.querySelector(".search-results");

    if (searchForm && searchInput) {
      // Handle search form submission
      searchForm.addEventListener("submit", function (e) {
        e.preventDefault();
        const query = searchInput.value.trim();
        if (query) {
          window.location.href = `/search?q=${encodeURIComponent(query)}`;
        }
      });

      // Keyboard navigation for search
      searchInput.addEventListener("keydown", function (e) {
        if (e.key === "Escape") {
          this.blur();
        }
      });
    }
  }

  /**
   * Contact form functionality - Handled by contact.js
   * Form validation functions removed to prevent duplication
   */

  /**
   * Theme toggle functionality with error boundary
   *
   * Two independent axes:
   * 1. Color theme: data-color-theme attribute (server-rendered from BLOG_THEME config)
   * 2. Light/dark mode: data-theme attribute ("dark", "light", or absent for auto)
   */
  function initThemeToggle() {
    var themeToggle = document.querySelector(".theme-toggle");
    if (!themeToggle) return;

    try {
      var savedMode = null;
      try {
        savedMode = localStorage.getItem("theme");
      } catch (localStorageError) {
        console.warn("localStorage access failed:", localStorageError);
      }

      applyTheme(savedMode);
      updateThemeToggle(resolveMode(savedMode));

      themeToggle.addEventListener("click", function () {
        try {
          var currentMode = getCurrentMode();
          var newMode = currentMode === "dark" ? "light" : "dark";

          try {
            localStorage.setItem("theme", newMode);
          } catch (storageError) {
            console.warn("Failed to save theme preference:", storageError);
          }

          applyTheme(newMode);
          updateThemeToggle(newMode);
        } catch (toggleError) {
          console.error("Theme toggle failed:", toggleError);
        }
      });

      // Respond to system preference changes when no explicit choice is saved
      if (window.matchMedia) {
        window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", function () {
          try {
            var saved = localStorage.getItem("theme");
            if (!saved) {
              applyTheme(null);
              updateThemeToggle(getCurrentMode());
            }
          } catch (e) {
            // ignore — localStorage or matchMedia failure
          }
        });
      }
    } catch (error) {
      console.error("Theme toggle initialization failed:", error);
    }
  }

  /**
   * Apply light/dark mode via data-theme attribute.
   * Color theme is handled separately via data-color-theme (server-rendered).
   */
  function applyTheme(mode) {
    if (mode === "dark" || mode === "light") {
      document.documentElement.setAttribute("data-theme", mode);
    } else {
      document.documentElement.removeAttribute("data-theme");
    }
  }

  /**
   * Resolve an explicit mode value to "dark" or "light", falling back to system preference.
   */
  function resolveMode(mode) {
    if (mode === "dark" || mode === "light") return mode;
    if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
      return "dark";
    }
    return "light";
  }

  /**
   * Get the current effective mode by reading localStorage then system preference.
   */
  function getCurrentMode() {
    try {
      var saved = localStorage.getItem("theme");
      if (saved === "dark" || saved === "light") return saved;
    } catch (e) {
      // ignore
    }
    return resolveMode(null);
  }

  /**
   * Update theme toggle button icon
   */
  function updateThemeToggle(theme) {
    var themeToggle = document.querySelector(".theme-toggle");
    if (!themeToggle) return;

    var icon = themeToggle.querySelector("svg");
    if (!icon) return;

    // SVG path constants — hardcoded, safe for innerHTML
    var sunPath = '<path d="M8 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM8 0a.5.5 0 0 1 .5.5v2a.5.5 0 0 1-1 0v-2A.5.5 0 0 1 8 0zm0 13a.5.5 0 0 1 .5.5v2a.5.5 0 0 1-1 0v-2A.5.5 0 0 1 8 13zm8-5a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1 0-1h2a.5.5 0 0 1 .5.5zM3 8a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1 0-1h2A.5.5 0 0 1 3 8zm10.657-5.657a.5.5 0 0 1 0 .707l-1.414 1.415a.5.5 0 1 1-.707-.708l1.414-1.414a.5.5 0 0 1 .707 0zm-9.193 9.193a.5.5 0 0 1 0 .707L3.05 13.657a.5.5 0 0 1-.707-.707l1.414-1.414a.5.5 0 0 1 .707 0zm9.193 2.121a.5.5 0 0 1-.707 0l-1.414-1.414a.5.5 0 0 1 .707-.707l1.414 1.414a.5.5 0 0 1 0 .707zM4.464 4.465a.5.5 0 0 1-.707 0L2.343 3.05a.5.5 0 1 1 .707-.707l1.414 1.414a.5.5 0 0 1 0 .708z"/>';
    var moonPath = '<path d="M6 .278a.768.768 0 0 1 .08.858 7.208 7.208 0 0 0-.878 3.46c0 4.021 3.278 7.277 7.318 7.277.527 0 1.04-.055 1.533-.16a.787.787 0 0 1 .81.316.733.733 0 0 1-.031.893A8.349 8.349 0 0 1 8.344 16C3.734 16 0 12.286 0 7.71 0 4.266 2.114 1.312 5.124.06A.752.752 0 0 1 6 .278z"/>';

    if (theme === "dark") {
      icon.innerHTML = sunPath;
      themeToggle.setAttribute("aria-label", "Switch to light mode");
    } else {
      icon.innerHTML = moonPath;
      themeToggle.setAttribute("aria-label", "Switch to dark mode");
    }
  }

  /**
   * Login — handles both the nav popover (public pages) and inline auth gate (protected pages).
   * Both forms share the .login-form class and POST to /login via fetch.
   */
  function initLoginPopover() {
    // Attach fetch-based submit to all login forms (popover + auth gate)
    var forms = document.querySelectorAll(".login-form");
    forms.forEach(function (form) {
      form.addEventListener("submit", function (e) {
        e.preventDefault();
        var errorEl = form.querySelector(".login-error");
        if (errorEl) {
          errorEl.hidden = true;
          errorEl.textContent = "";
        }

        var submitBtn = form.querySelector('button[type="submit"]');
        if (submitBtn) submitBtn.disabled = true;

        fetch("/login", {
          method: "POST",
          headers: { Accept: "application/json" },
          body: new FormData(form),
          credentials: "same-origin",
        })
          .then(function (res) {
            if (res.status === 403) {
              throw new Error("Session expired. Please refresh the page and try again.");
            }
            if (res.status === 429) {
              throw new Error("Too many attempts. Please wait and try again.");
            }
            return res.json().then(
              function (data) { return { ok: res.ok, data: data }; },
              function () { throw new Error("Server error. Please refresh the page."); }
            );
          })
          .then(function (result) {
            if (result.data.success) {
              window.location.href = result.data.redirect || window.location.pathname;
            } else {
              if (errorEl) {
                errorEl.textContent = result.data.error || "Login failed.";
                errorEl.hidden = false;
              }
              if (submitBtn) submitBtn.disabled = false;
            }
          })
          .catch(function (err) {
            if (errorEl) {
              errorEl.textContent = err.message || "Network error. Please try again.";
              errorEl.hidden = false;
            }
            if (submitBtn) submitBtn.disabled = false;
          });
      });
    });

    // Auto-focus the inline auth gate form (protected pages)
    var authGateInput = document.querySelector(".auth-gate-form input[name='username']");
    if (authGateInput) {
      authGateInput.focus();
    }

    // Nav popover toggle (public pages only)
    var popover = document.getElementById("login-popover");
    var trigger = document.querySelector(".login-trigger");
    if (!popover || !trigger) return;

    function openPopover() {
      popover.hidden = false;
      var firstInput = popover.querySelector('input[name="username"]');
      if (firstInput) firstInput.focus();
    }

    function closePopover() {
      popover.hidden = true;
      var errorEl = popover.querySelector(".login-error");
      if (errorEl) {
        errorEl.hidden = true;
        errorEl.textContent = "";
      }
    }

    trigger.addEventListener("click", function (e) {
      e.stopPropagation();
      if (popover.hidden) {
        openPopover();
      } else {
        closePopover();
      }
    });

    document.addEventListener("keydown", function (e) {
      if (e.key === "Escape" && !popover.hidden) {
        closePopover();
      }
    });

    document.addEventListener("click", function (e) {
      if (!popover.hidden && !popover.contains(e.target) && e.target !== trigger) {
        closePopover();
      }
    });

    popover.addEventListener("click", function (e) {
      e.stopPropagation();
    });
  }

  /**
   * Scroll behavior
   */
  function initScrollBehavior() {
    // Back to top button
    const backToTopBtn = document.createElement("button");
    backToTopBtn.className = "back-to-top";
    backToTopBtn.innerHTML = "↑";
    backToTopBtn.title = "Back to top";
    backToTopBtn.style.display = "none";
    backToTopBtn.setAttribute("aria-label", "Back to top");
    document.body.appendChild(backToTopBtn);

    // Simple scroll handler for back-to-top button
    const handleScroll = function () {
      const currentScrollY = window.scrollY;

      // Show/hide back to top button
      if (currentScrollY > 500) {
        backToTopBtn.style.display = "block";
      } else {
        backToTopBtn.style.display = "none";
      }
    };

    window.addEventListener("scroll", handleScroll, { passive: true });

    // Smooth scroll to top
    backToTopBtn.addEventListener("click", function () {
      window.scrollTo({
        top: 0,
        behavior: "smooth",
      });
    });

    // Simple smooth scroll for anchor links
    document.addEventListener("click", function (e) {
      if (e.target.matches('a[href^="#"]')) {
        e.preventDefault();
        const target = document.querySelector(e.target.getAttribute("href"));
        if (target) {
          target.scrollIntoView({
            behavior: "smooth",
            block: "start",
          });
        }
      }
    });
  }

  /**
   * Lazy loading for images with error boundary
   */
  function initLazyLoading() {
    try {
      if ("IntersectionObserver" in window) {
        const imageObserver = new IntersectionObserver((entries, observer) => {
          entries.forEach((entry) => {
            try {
              if (entry.isIntersecting) {
                const img = entry.target;
                
                // Add error handling for image loading
                img.onerror = function() {
                  console.warn("Failed to load lazy image:", img.dataset.src);
                  img.classList.add("lazy-error");
                  // Optionally set a fallback image
                  // img.src = "/static/img/placeholder.jpg";
                };
                
                img.onload = function() {
                  img.classList.remove("lazy");
                  img.classList.add("lazy-loaded");
                };
                
                img.src = img.dataset.src;
                imageObserver.unobserve(img);
              }
            } catch (entryError) {
              console.error("Lazy loading entry processing failed:", entryError);
              // Continue processing other entries
            }
          });
        });

        const lazyImages = document.querySelectorAll("img[data-src]");
        lazyImages.forEach((img) => {
          try {
            img.classList.add("lazy");
            imageObserver.observe(img);
          } catch (observeError) {
            console.error("Failed to observe lazy image:", observeError);
            // Fallback: load image immediately
            img.src = img.dataset.src;
          }
        });
      } else {
        // Fallback for browsers without IntersectionObserver
        console.info("IntersectionObserver not supported, loading all images immediately");
        const lazyImages = document.querySelectorAll("img[data-src]");
        lazyImages.forEach((img) => {
          img.src = img.dataset.src;
          img.classList.remove("lazy");
        });
      }
    } catch (error) {
      console.error("Lazy loading initialization failed:", error);
      // Fallback: load all images immediately
      try {
        const lazyImages = document.querySelectorAll("img[data-src]");
        lazyImages.forEach((img) => {
          img.src = img.dataset.src;
        });
      } catch (fallbackError) {
        console.error("Lazy loading fallback failed:", fallbackError);
      }
    }
  }

  /**
   * Utility Functions
   */

  /**
   * Copy text to clipboard
   */
  function copyToClipboard(text) {
    if (navigator.clipboard && window.isSecureContext) {
      return navigator.clipboard.writeText(text);
    } else {
      // Fallback for older browsers
      return new Promise((resolve, reject) => {
        const textArea = document.createElement("textarea");
        textArea.value = text;
        textArea.style.position = "fixed";
        textArea.style.left = "-999999px";
        textArea.style.top = "-999999px";
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();

        try {
          document.execCommand("copy");
          textArea.remove();
          resolve();
        } catch (err) {
          textArea.remove();
          reject(err);
        }
      });
    }
  }

})();
