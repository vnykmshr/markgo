// Debug page auto-refresh functionality
let refreshInterval = null;
let isRefreshing = false;

function startAutoRefresh() {
    refreshInterval = setInterval(() => {
        if (!isRefreshing) {
            refreshPage();
        }
    }, 30000); // Refresh every 30 seconds
}

function stopAutoRefresh() {
    if (refreshInterval) {
        clearInterval(refreshInterval);
        refreshInterval = null;
    }
}

function refreshPage() {
    isRefreshing = true;
    window.location.reload();
}

function hideRefreshIndicator(indicator) {
    indicator.style.display = 'none';
}

// Initialize auto-refresh functionality
document.addEventListener('DOMContentLoaded', () => {
    startAutoRefresh();
    
    // Create refresh indicator using DOM methods
    const refreshIndicator = document.createElement('div');
    refreshIndicator.className = 'refresh-indicator';
    
    const refreshContent = document.createElement('div');
    refreshContent.className = 'refresh-content';
    
    const refreshSpinner = document.createElement('div');
    refreshSpinner.className = 'refresh-spinner';
    
    const refreshText = document.createElement('span');
    refreshText.textContent = 'Auto-refreshing every 30s';
    
    const refreshClose = document.createElement('button');
    refreshClose.className = 'refresh-close';
    refreshClose.textContent = '×';
    
    // Add event listener instead of onclick
    refreshClose.addEventListener('click', () => {
        stopAutoRefresh();
        hideRefreshIndicator(refreshIndicator);
    });
    
    refreshContent.appendChild(refreshSpinner);
    refreshContent.appendChild(refreshText);
    refreshContent.appendChild(refreshClose);
    refreshIndicator.appendChild(refreshContent);
    
    document.body.appendChild(refreshIndicator);
});

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    stopAutoRefresh();
});