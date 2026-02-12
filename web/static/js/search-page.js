/**
 * Search Page — auto-focus, empty query prevention, result highlighting.
 */

function escapeRegex(string) {
    return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function highlightInElement(element, term) {
    const text = element.textContent;
    const regex = new RegExp(`(${escapeRegex(term)})`, 'gi');
    const parts = text.split(regex);

    if (parts.length <= 1) return; // no matches

    element.textContent = '';
    parts.forEach((part) => {
        if (regex.test(part)) {
            // Reset regex lastIndex after test
            regex.lastIndex = 0;
            const mark = document.createElement('mark');
            mark.textContent = part;
            element.appendChild(mark);
        } else {
            element.appendChild(document.createTextNode(part));
        }
    });
}

function highlightSearchTerms(term) {
    document.querySelectorAll('.search-result').forEach((article) => {
        const title = article.querySelector('.result-title a');
        const excerpt = article.querySelector('.result-excerpt');

        if (title) highlightInElement(title, term);
        if (excerpt) highlightInElement(excerpt, term);
    });
}

export function init() {
    const searchInput = document.querySelector('.search-input');
    const searchForm = document.querySelector('.search-form');

    // Auto-focus search input on desktop
    if (searchInput && window.innerWidth > 768) {
        searchInput.focus();
    }

    // Prevent empty submissions
    if (searchForm) {
        searchForm.addEventListener('submit', (e) => {
            const query = searchInput.value.trim();
            if (!query) {
                e.preventDefault();
                searchInput.focus();
            }
        });
    }

    // Highlight search terms in results
    const searchTerm = new URLSearchParams(window.location.search).get('q');
    if (searchTerm) {
        highlightSearchTerms(searchTerm);
    }
}
