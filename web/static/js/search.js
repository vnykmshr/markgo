/* Search Page JavaScript */

document.addEventListener('DOMContentLoaded', function() {
    const searchInput = document.querySelector('.search-input');
    const searchForm = document.querySelector('.search-form');

    // Auto-focus search input on page load (except on mobile)
    if (searchInput && window.innerWidth > 768) {
        searchInput.focus();
    }

    // Handle search form submission
    if (searchForm) {
        searchForm.addEventListener('submit', function(e) {
            const query = searchInput.value.trim();
            if (!query) {
                e.preventDefault();
                searchInput.focus();
                return;
            }
        });
    }

    // Real-time search suggestions (if implemented)
    if (searchInput) {
        let searchTimeout;
        searchInput.addEventListener('input', function() {
            clearTimeout(searchTimeout);
            const query = this.value.trim();

            if (query.length < 2) {
                return;
            }

            searchTimeout = setTimeout(() => {
                // Implement search suggestions here if needed
                console.log('Searching for:', query);
            }, 300);
        });
    }

    // Highlight search terms in results
    const searchTerm = new URLSearchParams(window.location.search).get('q');
    if (searchTerm) {
        highlightSearchTerms(searchTerm);
    }

    function highlightSearchTerms(term) {
        const articles = document.querySelectorAll('.search-result');
        articles.forEach(article => {
            const title = article.querySelector('.result-title a');
            const excerpt = article.querySelector('.result-excerpt');

            if (title) {
                title.innerHTML = highlightText(title.textContent, term);
            }
            if (excerpt) {
                excerpt.innerHTML = highlightText(excerpt.textContent, term);
            }
        });
    }

    function highlightText(text, term) {
        const regex = new RegExp(`(${escapeRegex(term)})`, 'gi');
        return text.replace(regex, '<mark>$1</mark>');
    }

    function escapeRegex(string) {
        return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    }
});
