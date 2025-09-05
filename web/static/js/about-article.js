/**
 * About Article Page JavaScript
 * Handles about page functionality including animations, clipboard copying, and interactive elements
 * @fileoverview About page functionality for MarkGo blog engine
 */

/**
 * Copies text to clipboard with fallback support for older browsers
 * @param {string} text - Text content to copy to clipboard
 * @global
 */
function copyToClipboard(text) {
    if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(text).then(() => {
            // Visual feedback
            const button = event.target.closest('.copy-link');
            const originalText = button.innerHTML;
            button.innerHTML = '<svg width="20" height="20" fill="currentColor" viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>Copied!';
            setTimeout(() => {
                button.innerHTML = originalText;
            }, 2000);
        }).catch(err => {
            console.error('Failed to copy: ', err);
            fallbackCopyTextToClipboard(text);
        });
    } else {
        fallbackCopyTextToClipboard(text);
    }
}

/**
 * Fallback clipboard copy method for older browsers
 * Uses the deprecated document.execCommand method as last resort
 * @param {string} text - Text content to copy to clipboard
 */
function fallbackCopyTextToClipboard(text) {
    const textArea = document.createElement("textarea");
    textArea.value = text;

    // Avoid scrolling to bottom
    textArea.style.top = "0";
    textArea.style.left = "0";
    textArea.style.position = "fixed";

    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    try {
        const successful = document.execCommand('copy');
        if (successful) {
            const button = event.target.closest('.copy-link');
            const originalText = button.innerHTML;
            button.innerHTML = '<svg width="20" height="20" fill="currentColor" viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>Copied!';
            setTimeout(() => {
                button.innerHTML = originalText;
            }, 2000);
        }
    } catch (err) {
        console.error('Fallback: Oops, unable to copy', err);
    }

    document.body.removeChild(textArea);
}

/**
 * Initialize about page functionality
 * Sets up animations, interactions, and scroll observers
 */
document.addEventListener('DOMContentLoaded', function() {
    initializeSkillBars();
    initializeSmoothScrolling();
    initializeTimelineAnimations();
    initializeSocialLinkEffects();

    /**
     * Initializes skill progress bar animations
     * Uses Intersection Observer to trigger animations when bars come into view
     */
    function initializeSkillBars() {
        const skillBars = document.querySelectorAll('.skill-progress');

        const observerOptions = {
            threshold: 0.5,
            rootMargin: '0px 0px -50px 0px'
        };

        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const progressBar = entry.target;
                    const targetWidth = progressBar.dataset.width || '0%';
                    progressBar.style.width = targetWidth;
                }
            });
        }, observerOptions);

        skillBars.forEach(bar => {
            observer.observe(bar);
        });
    }

    /**
     * Adds smooth scrolling behavior to anchor links
     * Prevents default jump behavior and uses smooth scroll instead
     */
    function initializeSmoothScrolling() {
        const anchorLinks = document.querySelectorAll('a[href^="#"]');
        anchorLinks.forEach(link => {
            link.addEventListener('click', function(e) {
                e.preventDefault();
                const targetId = this.getAttribute('href');
                const targetElement = document.querySelector(targetId);

                if (targetElement) {
                    targetElement.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            });
        });
    }

    /**
     * Initializes timeline item animations on scroll
     * Items fade in and slide from left when they come into view
     */
    function initializeTimelineAnimations() {
        const timelineItems = document.querySelectorAll('.timeline-item');
        const timelineObserver = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    entry.target.style.opacity = '1';
                    entry.target.style.transform = 'translateX(0)';
                }
            });
        }, {
            threshold: 0.3,
            rootMargin: '0px 0px -100px 0px'
        });

        timelineItems.forEach(item => {
            item.style.opacity = '0';
            item.style.transform = 'translateX(-20px)';
            item.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
            timelineObserver.observe(item);
        });
    }

    /**
     * Adds hover effects to social media links
     * Creates subtle scale and translate animations
     */
    function initializeSocialLinkEffects() {
        const socialLinks = document.querySelectorAll('.social-link');
        socialLinks.forEach(link => {
            link.addEventListener('mouseenter', function() {
                this.style.transform = 'translateY(-2px) scale(1.05)';
            });

            link.addEventListener('mouseleave', function() {
                this.style.transform = 'translateY(-2px) scale(1)';
            });
        });
    }
});
