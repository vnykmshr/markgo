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
    initScrollBehavior();
    initLazyLoading();
    initAnalytics();
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
        (currentPath.startsWith("/articles/") && linkPath === "/articles") ||
        (currentPath.startsWith("/tags/") && linkPath === "/tags") ||
        (currentPath.startsWith("/categories/") && linkPath === "/categories")
      ) {
        link.classList.add("active");
      }
    });

    // Simple navbar scroll behavior - just add shadow
    const handleNavbarScroll = throttle(function () {
      const currentScrollY = window.scrollY;

      if (navbar) {
        // Add shadow when scrolled
        if (currentScrollY > 50) {
          navbar.classList.add("scrolled");
        } else {
          navbar.classList.remove("scrolled");
        }
      }
    }, 16);

    window.addEventListener("scroll", handleNavbarScroll, { passive: true });
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
      // Real-time search suggestions
      let searchTimeout;

      searchInput.addEventListener("input", function () {
        clearTimeout(searchTimeout);
        const query = this.value.trim();

        if (query.length > 2) {
          searchTimeout = setTimeout(function () {
            fetchSearchSuggestions(query);
          }, 300);
        } else {
          hideSearchSuggestions();
        }
      });

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
          hideSearchSuggestions();
          this.blur();
        }
      });

      // Close suggestions when clicking outside
      document.addEventListener("click", function (e) {
        if (!e.target.closest(".search-container")) {
          hideSearchSuggestions();
        }
      });
    }
  }

  /**
   * Fetch search suggestions with error boundary
   */
  function fetchSearchSuggestions(query) {
    try {
      // TODO: Implement actual search suggestions API call
      // For now, we'll implement basic client-side filtering with error boundary
      console.log("Searching for:", query);

      // Future API implementation with proper error boundary:
      /*
      fetch(`/api/search/suggestions?q=${encodeURIComponent(query)}`)
        .then(response => {
          if (!response.ok) {
            throw new Error(`Search API failed: ${response.status}`);
          }
          return response.json();
        })
        .then(data => showSearchSuggestions(data))
        .catch(err => {
          console.error('Search suggestions failed:', err);
          // Graceful fallback - hide suggestions instead of crashing
          hideSearchSuggestions();
          // Optional: Show user-friendly message
          // showMessage('Search suggestions temporarily unavailable', 'warning');
        });
      */
    } catch (error) {
      console.error('Search suggestions error:', error);
      hideSearchSuggestions();
    }
  }

  /**
   * Show search suggestions
   */
  function showSearchSuggestions(suggestions) {
    const searchContainer = document.querySelector(".search-container");
    if (!searchContainer) return;

    let suggestionsEl = searchContainer.querySelector(".search-suggestions");
    if (!suggestionsEl) {
      suggestionsEl = document.createElement("div");
      suggestionsEl.className = "search-suggestions";
      searchContainer.appendChild(suggestionsEl);
    }

    if (suggestions && suggestions.length > 0) {
      suggestionsEl.innerHTML = suggestions
        .map(
          (suggestion) =>
            `<a href="/search?q=${encodeURIComponent(suggestion)}" class="suggestion-item">${suggestion}</a>`,
        )
        .join("");
      suggestionsEl.style.display = "block";
    } else {
      suggestionsEl.style.display = "none";
    }
  }

  /**
   * Hide search suggestions
   */
  function hideSearchSuggestions() {
    const suggestions = document.querySelector(".search-suggestions");
    if (suggestions) {
      suggestions.style.display = "none";
    }
  }

  /**
   * Contact form functionality - Handled by contact.js
   * Form validation functions removed to prevent duplication
   */

  /**
   * Theme toggle functionality with error boundary
   */
  function initThemeToggle() {
    const themeToggle = document.querySelector(".theme-toggle");
    if (!themeToggle) return;

    try {
      // Get saved theme or default to 'light' with error boundary
      let savedTheme = "light";
      try {
        savedTheme = localStorage.getItem("theme") || "light";
      } catch (localStorageError) {
        console.warn("localStorage access failed, using default theme:", localStorageError);
        // Fallback to system preference if localStorage fails
        if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
          savedTheme = "dark";
        }
      }

      document.documentElement.setAttribute("data-theme", savedTheme);
      updateThemeToggle(savedTheme);

      themeToggle.addEventListener("click", function () {
        try {
          const currentTheme = document.documentElement.getAttribute("data-theme");
          const newTheme = currentTheme === "dark" ? "light" : "dark";

          document.documentElement.setAttribute("data-theme", newTheme);
          
          // Safe localStorage access
          try {
            localStorage.setItem("theme", newTheme);
          } catch (storageError) {
            console.warn("Failed to save theme preference:", storageError);
            // Theme still works, just won't persist
          }
          
          updateThemeToggle(newTheme);
        } catch (toggleError) {
          console.error("Theme toggle failed:", toggleError);
          // Prevent theme toggle from crashing the app
        }
      });
    } catch (error) {
      console.error("Theme toggle initialization failed:", error);
      // App continues to work without theme toggle
    }
  }

  /**
   * Update theme toggle button
   */
  function updateThemeToggle(theme) {
    const themeToggle = document.querySelector(".theme-toggle");
    if (!themeToggle) return;

    const icon = themeToggle.querySelector("svg");
    if (theme === "dark") {
      icon.innerHTML =
        '<path d="M8 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM8 0a.5.5 0 0 1 .5.5v2a.5.5 0 0 1-1 0v-2A.5.5 0 0 1 8 0zm0 13a.5.5 0 0 1 .5.5v2a.5.5 0 0 1-1 0v-2A.5.5 0 0 1 8 13zm8-5a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1 0-1h2a.5.5 0 0 1 .5.5zM3 8a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1 0-1h2A.5.5 0 0 1 3 8zm10.657-5.657a.5.5 0 0 1 0 .707l-1.414 1.415a.5.5 0 1 1-.707-.708l1.414-1.414a.5.5 0 0 1 .707 0zm-9.193 9.193a.5.5 0 0 1 0 .707L3.05 13.657a.5.5 0 0 1-.707-.707l1.414-1.414a.5.5 0 0 1 .707 0zm9.193 2.121a.5.5 0 0 1-.707 0l-1.414-1.414a.5.5 0 0 1 .707-.707l1.414 1.414a.5.5 0 0 1 0 .707zM4.464 4.465a.5.5 0 0 1-.707 0L2.343 3.05a.5.5 0 1 1 .707-.707l1.414 1.414a.5.5 0 0 1 0 .708z"/>';
    } else {
      icon.innerHTML =
        '<path d="M6 .278a.768.768 0 0 1 .08.858 7.208 7.208 0 0 0-.878 3.46c0 4.021 3.278 7.277 7.318 7.277.527 0 1.04-.055 1.533-.16a.787.787 0 0 1 .81.316.733.733 0 0 1-.031.893A8.349 8.349 0 0 1 8.344 16C3.734 16 0 12.286 0 7.71 0 4.266 2.114 1.312 5.124.06A.752.752 0 0 1 6 .278z"/>';
    }
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
   * Analytics initialization with error boundary
   */
  function initAnalytics() {
    try {
      // Simple page view tracking with error boundary
      if (typeof gtag !== "undefined") {
        try {
          gtag("config", "GA_MEASUREMENT_ID", {
            page_title: document.title,
            page_location: window.location.href,
          });
        } catch (gtagError) {
          console.error("Google Analytics configuration failed:", gtagError);
        }
      }

      // Track outbound links with error boundary
      document.addEventListener("click", function (e) {
        try {
          if (
            e.target.matches('a[href^="http"]') &&
            !e.target.href.includes(window.location.hostname)
          ) {
            if (typeof gtag !== "undefined") {
              try {
                gtag("event", "click", {
                  event_category: "outbound",
                  event_label: e.target.href,
                });
              } catch (eventError) {
                console.error("Analytics event tracking failed:", eventError);
              }
            }
          }
        } catch (clickError) {
          console.error("Outbound link tracking failed:", clickError);
          // Don't prevent the link from working
        }
      });
    } catch (error) {
      console.error("Analytics initialization failed:", error);
      // App continues to work without analytics
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

  /**
   * Show message to user
   */
  function showMessage(text, type = "info") {
    const message = document.createElement("div");
    message.className = `message message-${type}`;
    message.textContent = text;

    // Style the message
    Object.assign(message.style, {
      position: "fixed",
      top: "20px",
      right: "20px",
      padding: "12px 20px",
      borderRadius: "8px",
      color: "white",
      fontWeight: "500",
      zIndex: "9999",
      maxWidth: "400px",
      opacity: "0",
      transform: "translateX(100%)",
      transition: "all 0.3s ease",
    });

    // Set background color based on type
    switch (type) {
      case "success":
        message.style.backgroundColor = "#10b981";
        break;
      case "error":
        message.style.backgroundColor = "#ef4444";
        break;
      case "warning":
        message.style.backgroundColor = "#f59e0b";
        break;
      default:
        message.style.backgroundColor = "#3b82f6";
    }

    document.body.appendChild(message);

    // Animate in
    setTimeout(() => {
      message.style.opacity = "1";
      message.style.transform = "translateX(0)";
    }, 100);

    // Remove after 5 seconds
    setTimeout(() => {
      message.style.opacity = "0";
      message.style.transform = "translateX(100%)";
      setTimeout(() => {
        if (message.parentElement) {
          message.parentElement.removeChild(message);
        }
      }, 300);
    }, 5000);
  }

  /**
   * Debounce function
   */
  function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
      const later = () => {
        clearTimeout(timeout);
        func(...args);
      };
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
    };
  }

  /**
   * Throttle function
   */
  function throttle(func, limit) {
    let inThrottle;
    return function () {
      const args = arguments;
      const context = this;
      if (!inThrottle) {
        func.apply(context, args);
        inThrottle = true;
        setTimeout(() => (inThrottle = false), limit);
      }
    };
  }

  // Make utility functions globally available
  window.markgo = {
    copyToClipboard,
    showMessage,
    debounce,
    throttle,
  };
})();
