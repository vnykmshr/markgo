// Admin Dashboard JavaScript

// Notification system
function showNotification(title, message, type = 'success') {
    const container = document.getElementById('notification-container');
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    
    // Create close button
    const closeBtn = document.createElement('button');
    closeBtn.className = 'notification-close';
    closeBtn.textContent = '×';
    notification.appendChild(closeBtn);
    
    // Create title element
    const titleElement = document.createElement('div');
    titleElement.className = 'notification-title';
    titleElement.textContent = title;
    notification.appendChild(titleElement);
    
    // Create message element
    const messageElement = document.createElement('div');
    messageElement.className = 'notification-message';
    messageElement.textContent = message;
    notification.appendChild(messageElement);
    
    container.appendChild(notification);
    
    // Add event listener to the close button
    closeBtn.addEventListener('click', function() {
        removeNotification(this);
    });
    
    // Auto-remove after 5 seconds
    setTimeout(() => {
        if (notification.parentNode) {
            removeNotification(notification.querySelector('.notification-close'));
        }
    }, 5000);
}

function removeNotification(button) {
    const notification = button.parentNode;
    notification.className += ' notification-slide-out';
    setTimeout(() => {
        if (notification.parentNode) {
            notification.parentNode.removeChild(notification);
        }
    }, 300);
}

// Modal functions  
function showModal(title, message, data = null, isSuccess = true) {
    document.getElementById('modal-title').textContent = title;
    
    // Use innerHTML carefully - ensure no script content
    const messageElement = document.getElementById('modal-message');
    messageElement.innerHTML = message;
    
    showModalCommon(data, isSuccess);
}

function showModalWithElement(title, messageElement, data = null, isSuccess = true) {
    document.getElementById('modal-title').textContent = title;
    
    // Clear and append DOM element safely
    const messageContainer = document.getElementById('modal-message');
    messageContainer.innerHTML = ''; // Clear existing content
    messageContainer.appendChild(messageElement);
    
    showModalCommon(data, isSuccess);
}

function showModalCommon(data = null, isSuccess = true) {
    const dataElement = document.getElementById('modal-data');
    if (data && typeof data === 'object') {
        dataElement.textContent = JSON.stringify(data, null, 2);
        dataElement.className = 'modal-data-visible';
    } else if (data && typeof data === 'string') {
        dataElement.textContent = data;
        dataElement.className = 'modal-data-visible';
    } else {
        dataElement.className = 'modal-data-hidden';
    }
    
    // Change header color based on success/error
    const header = document.querySelector('.modal-header');
    // Remove existing color classes
    header.classList.remove('modal-header-success', 'modal-header-error', 'modal-header-default');
    if (isSuccess) {
        header.classList.add('modal-header-success');
    } else {
        header.classList.add('modal-header-error');
    }
    
    const modal = document.getElementById('result-modal');
    modal.className = 'modal modal-visible';
    document.body.classList.add('body-modal-open');
    document.body.classList.remove('body-modal-closed');
}

function closeModal() {
    const modal = document.getElementById('result-modal');
    modal.className = 'modal modal-hidden';
    document.body.classList.add('body-modal-closed');
    document.body.classList.remove('body-modal-open');
}

// GET action execution (for viewing data)
function executeGetAction(url, actionName) {
    // Find the button that was clicked
    const button = document.querySelector(`.admin-get-btn[data-url="${url}"][data-name="${actionName}"]`);
    if (!button) return;
    
    const originalText = button.textContent;
    button.textContent = 'Loading...';
    button.disabled = true;
    
    fetch(url, {
        method: 'GET',
        headers: {
            'Accept': 'application/json'
        }
    })
    .then(response => {
        if (response.ok) {
            return response.json();
        } else {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
    })
    .then(data => {
        // Create a more user-friendly message using DOM methods instead of template literals
        const messageContainer = document.createElement('div');
        
        // Add success message
        const successMsg = document.createElement('p');
        const strong = document.createElement('strong');
        strong.textContent = 'Data retrieved successfully';
        successMsg.appendChild(strong);
        messageContainer.appendChild(successMsg);
        
        // Add summary information if available
        if (data.drafts && Array.isArray(data.drafts)) {
            const countMsg = document.createElement('p');
            countMsg.textContent = `Found ${data.drafts.length} draft articles`;
            messageContainer.appendChild(countMsg);
            
            // Create a table for draft articles
            if (data.drafts.length > 0) {
                const tableContainer = document.createElement('div');
                tableContainer.className = 'draft-table-container';
                
                const table = document.createElement('table');
                table.className = 'draft-table';
                
                // Create table header
                const thead = document.createElement('thead');
                const headerRow = document.createElement('tr');
                ['Title', 'Author', 'Date', 'Categories'].forEach(headerText => {
                    const th = document.createElement('th');
                    th.textContent = headerText;
                    headerRow.appendChild(th);
                });
                thead.appendChild(headerRow);
                table.appendChild(thead);
                
                // Create table body
                const tbody = document.createElement('tbody');
                data.drafts.forEach(draft => {
                    const row = document.createElement('tr');
                    
                    const titleCell = document.createElement('td');
                    titleCell.className = 'title-cell';
                    titleCell.textContent = draft.title || 'Untitled';
                    titleCell.setAttribute('title', draft.title || 'Untitled');
                    row.appendChild(titleCell);
                    
                    const authorCell = document.createElement('td');
                    authorCell.textContent = draft.author || 'Unknown';
                    row.appendChild(authorCell);
                    
                    const dateCell = document.createElement('td');
                    dateCell.className = 'date-cell';
                    dateCell.textContent = new Date(draft.date).toLocaleDateString();
                    row.appendChild(dateCell);
                    
                    const categoriesCell = document.createElement('td');
                    categoriesCell.className = 'categories-cell';
                    const categories = draft.categories ? draft.categories.join(', ') : 'None';
                    categoriesCell.textContent = categories;
                    categoriesCell.setAttribute('title', categories);
                    row.appendChild(categoriesCell);
                    
                    tbody.appendChild(row);
                });
                table.appendChild(tbody);
                tableContainer.appendChild(table);
                messageContainer.appendChild(tableContainer);
            } else {
                const emptyMsg = document.createElement('p');
                const em = document.createElement('em');
                em.textContent = 'No draft articles found.';
                emptyMsg.appendChild(em);
                messageContainer.appendChild(emptyMsg);
            }
        } else if (data.articles) {
            const total = data.articles.total || 0;
            const published = data.articles.published || 0;
            const drafts = data.articles.drafts || 0;
            const statsMsg = document.createElement('p');
            statsMsg.textContent = `Total: ${total} articles (${published} published, ${drafts} drafts)`;
            messageContainer.appendChild(statsMsg);
        } else if (data.status === 'healthy') {
            const statusMsg = document.createElement('p');
            statusMsg.textContent = 'System Status: ';
            const healthySpan = document.createElement('span');
            healthySpan.className = 'status-healthy';
            healthySpan.textContent = 'Healthy ✅';
            statusMsg.appendChild(healthySpan);
            messageContainer.appendChild(statusMsg);
            
            if (data.uptime) {
                const uptimeMsg = document.createElement('p');
                uptimeMsg.textContent = `Uptime: ${data.uptime}`;
                messageContainer.appendChild(uptimeMsg);
            }
        } else if (data.memory) {
            const memMsg = document.createElement('p');
            memMsg.textContent = `Memory Allocated: ${data.memory.alloc || 'N/A'}`;
            messageContainer.appendChild(memMsg);
            
            if (data.uptime) {
                const uptimeMsg = document.createElement('p');
                uptimeMsg.textContent = `Uptime: ${data.uptime}`;
                messageContainer.appendChild(uptimeMsg);
            }
        }
        
        // Add timestamp
        const timestampMsg = document.createElement('p');
        const small = document.createElement('small');
        small.textContent = `Timestamp: ${new Date().toLocaleString()}`;
        timestampMsg.appendChild(small);
        messageContainer.appendChild(timestampMsg);
        
        // Show data in modal
        showModalWithElement(
            `📊 ${actionName}`,
            messageContainer,
            data,
            true
        );
        
        // Show notification for quick feedback
        showNotification(
            'Data Loaded',
            `${actionName} data loaded successfully`,
            'success'
        );
    })
    .catch(error => {
        console.error('Get action failed:', error);
        
        // Show error modal using DOM method
        const errorContainer = document.createElement('div');
        
        const errorTitle = document.createElement('p');
        const strongError = document.createElement('strong');
        strongError.textContent = 'Failed to load data:';
        errorTitle.appendChild(strongError);
        errorContainer.appendChild(errorTitle);
        
        const errorMsg = document.createElement('p');
        errorMsg.textContent = error.message || `Failed to load ${actionName}`;
        errorContainer.appendChild(errorMsg);
        
        const errorTimestamp = document.createElement('p');
        const smallError = document.createElement('small');
        smallError.textContent = `Timestamp: ${new Date().toLocaleString()}`;
        errorTimestamp.appendChild(smallError);
        errorContainer.appendChild(errorTimestamp);
        
        showModalWithElement(
            `❌ ${actionName} - Failed`,
            errorContainer,
            error.stack || error.toString(),
            false
        );
        
        // Also show notification
        showNotification(
            'Load Failed',
            error.message || `Failed to load ${actionName}`,
            'error'
        );
    })
    .finally(() => {
        button.textContent = originalText;
        button.disabled = false;
    });
}

// Admin action execution (for POST actions)
function executeAdminAction(url, actionName) {
    if (!confirm(`Are you sure you want to execute "${actionName}"?`)) {
        return;
    }
    
    // Find the button that was clicked
    const button = document.querySelector(`.admin-action-btn[data-url="${url}"][data-name="${actionName}"]`);
    if (!button) return;
    
    const originalText = button.textContent;
    button.textContent = 'Executing...';
    button.disabled = true;
    
    fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        }
    })
    .then(response => {
        if (response.ok) {
            return response.json();
        } else {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
    })
    .then(data => {
        // Show success modal with detailed results using DOM method
        const successContainer = document.createElement('div');
        
        const successMsg = document.createElement('p');
        const strongSuccess = document.createElement('strong');
        strongSuccess.textContent = data.message || `${actionName} completed successfully`;
        successMsg.appendChild(strongSuccess);
        successContainer.appendChild(successMsg);
        
        const successTimestamp = document.createElement('p');
        const smallSuccess = document.createElement('small');
        smallSuccess.textContent = `Timestamp: ${new Date().toLocaleString()}`;
        successTimestamp.appendChild(smallSuccess);
        successContainer.appendChild(successTimestamp);
        
        showModalWithElement(
            `✅ ${actionName} - Success`,
            successContainer,
            data,
            true
        );
        
        // Show notification for quick feedback
        showNotification(
            'Action Completed',
            message,
            'success'
        );
        
        // If it's a reload or cache clear action, offer to refresh the page
        if (url.includes('reload') || url.includes('cache/clear')) {
            setTimeout(() => {
                if (confirm('Action completed successfully. Refresh the page to see updated data?')) {
                    window.location.reload();
                }
            }, 2000);
        }
    })
    .catch(error => {
        console.error('Admin action failed:', error);
        
        // Show error modal using DOM method
        const errorContainer = document.createElement('div');
        
        const errorTitle = document.createElement('p');
        const strongError = document.createElement('strong');
        strongError.textContent = 'Action failed:';
        errorTitle.appendChild(strongError);
        errorContainer.appendChild(errorTitle);
        
        const errorMsg = document.createElement('p');
        errorMsg.textContent = error.message || `Failed to execute ${actionName}`;
        errorContainer.appendChild(errorMsg);
        
        const errorTimestamp = document.createElement('p');
        const smallError = document.createElement('small');
        smallError.textContent = `Timestamp: ${new Date().toLocaleString()}`;
        errorTimestamp.appendChild(smallError);
        errorContainer.appendChild(errorTimestamp);
        
        showModalWithElement(
            `❌ ${actionName} - Failed`,
            errorContainer,
            error.stack || error.toString(),
            false
        );
        
        // Also show notification
        showNotification(
            'Action Failed',
            error.message || `Failed to execute ${actionName}`,
            'error'
        );
    })
    .finally(() => {
        button.textContent = originalText;
        button.disabled = false;
    });
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    
    // Add event listeners for admin action buttons (POST)
    const adminButtons = document.querySelectorAll('.admin-action-btn');
    adminButtons.forEach(button => {
        button.addEventListener('click', function(event) {
            const url = this.dataset.url;
            const actionName = this.dataset.name;
            executeAdminAction(url, actionName);
        });
    });
    
    // Add event listeners for admin GET buttons 
    const getButtons = document.querySelectorAll('.admin-get-btn');
    getButtons.forEach(button => {
        button.addEventListener('click', function(event) {
            const url = this.dataset.url;
            const actionName = this.dataset.name;
            executeGetAction(url, actionName);
        });
    });
    
    // Add event listeners for modal close buttons
    const modalCloseBtns = document.querySelectorAll('#modal-close-btn, #modal-close-footer');
    modalCloseBtns.forEach(btn => {
        btn.addEventListener('click', closeModal);
    });
    
    // Close modal when clicking outside of it
    window.onclick = function(event) {
        const modal = document.getElementById('result-modal');
        if (event.target === modal) {
            closeModal();
        }
    }
    
    // Close modal with Escape key
    document.addEventListener('keydown', function(event) {
        if (event.key === 'Escape') {
            closeModal();
        }
    });
});