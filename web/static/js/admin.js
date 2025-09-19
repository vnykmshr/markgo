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
        } else if (data.stats) {
            // Handle preview sessions data
            const previewMsg = document.createElement('p');
            const strongPreview = document.createElement('strong');
            strongPreview.textContent = 'Preview Service Status';
            previewMsg.appendChild(strongPreview);
            messageContainer.appendChild(previewMsg);

            const serviceStatus = document.createElement('p');
            serviceStatus.textContent = `Service Running: ${data.service_running ? '✅ Yes' : '❌ No'}`;
            messageContainer.appendChild(serviceStatus);

            const activeSessions = document.createElement('p');
            activeSessions.textContent = `Active Sessions: ${data.stats.active_sessions || 0}`;
            messageContainer.appendChild(activeSessions);

            const totalClients = document.createElement('p');
            totalClients.textContent = `Connected Clients: ${data.stats.total_clients || 0}`;
            messageContainer.appendChild(totalClients);

            const filesWatched = document.createElement('p');
            filesWatched.textContent = `Files Watched: ${data.stats.files_watched || 0}`;
            messageContainer.appendChild(filesWatched);

            // Add actions for draft articles if we have any
            if (data.stats.active_sessions > 0) {
                const actionsMsg = document.createElement('p');
                const strongActions = document.createElement('strong');
                strongActions.textContent = 'Available Actions:';
                actionsMsg.appendChild(strongActions);
                messageContainer.appendChild(actionsMsg);

                const actionsList = document.createElement('ul');

                const createPreviewAction = document.createElement('li');
                const createBtn = document.createElement('button');
                createBtn.textContent = 'Create New Preview';
                createBtn.className = 'btn btn-primary create-preview-btn';
                createBtn.style.margin = '5px';
                createPreviewAction.appendChild(createBtn);
                actionsList.appendChild(createPreviewAction);

                messageContainer.appendChild(actionsList);
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

// Preview management functions
function createPreviewSession(articleSlug) {
    if (!articleSlug) {
        showNotification('Error', 'Article slug is required', 'error');
        return;
    }

    fetch('/api/preview/sessions', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        },
        body: JSON.stringify({ article_slug: articleSlug })
    })
    .then(response => {
        if (response.ok) {
            return response.json();
        } else {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
    })
    .then(data => {
        if (data.session) {
            // Show success modal with preview URL
            const successContainer = document.createElement('div');

            const successMsg = document.createElement('p');
            const strongSuccess = document.createElement('strong');
            strongSuccess.textContent = 'Preview session created successfully!';
            successMsg.appendChild(strongSuccess);
            successContainer.appendChild(successMsg);

            const urlMsg = document.createElement('p');
            urlMsg.textContent = 'Preview URL: ';
            const urlLink = document.createElement('a');
            urlLink.href = data.session.url;
            urlLink.target = '_blank';
            urlLink.textContent = data.session.url;
            urlLink.style.color = '#2196f3';
            urlMsg.appendChild(urlLink);
            successContainer.appendChild(urlMsg);

            const sessionInfo = document.createElement('p');
            sessionInfo.textContent = `Session ID: ${data.session.id}`;
            successContainer.appendChild(sessionInfo);

            const openBtn = document.createElement('button');
            openBtn.textContent = 'Open Preview';
            openBtn.className = 'btn btn-primary';
            openBtn.style.margin = '10px 5px 0 0';
            openBtn.onclick = () => window.open(data.session.url, '_blank');
            successContainer.appendChild(openBtn);

            showModalWithElement(
                '✅ Preview Session Created',
                successContainer,
                data,
                true
            );

            showNotification(
                'Preview Created',
                `Preview session created for ${articleSlug}`,
                'success'
            );
        }
    })
    .catch(error => {
        console.error('Failed to create preview session:', error);
        showNotification(
            'Preview Failed',
            error.message || 'Failed to create preview session',
            'error'
        );
    });
}

function showCreatePreviewDialog() {
    // First, fetch available drafts
    fetch('/admin/drafts', {
        method: 'GET',
        headers: {
            'Accept': 'application/json'
        }
    })
    .then(response => response.json())
    .then(data => {
        const dialogContainer = document.createElement('div');

        const titleMsg = document.createElement('p');
        const strongTitle = document.createElement('strong');
        strongTitle.textContent = 'Create Preview Session';
        titleMsg.appendChild(strongTitle);
        dialogContainer.appendChild(titleMsg);

        if (data.drafts && data.drafts.length > 0) {
            const selectMsg = document.createElement('p');
            selectMsg.textContent = 'Select a draft article to preview:';
            dialogContainer.appendChild(selectMsg);

            const select = document.createElement('select');
            select.id = 'preview-article-select';
            select.style.width = '100%';
            select.style.padding = '8px';
            select.style.marginBottom = '15px';

            const defaultOption = document.createElement('option');
            defaultOption.value = '';
            defaultOption.textContent = '-- Select an article --';
            select.appendChild(defaultOption);

            data.drafts.forEach(draft => {
                const option = document.createElement('option');
                option.value = draft.slug;
                option.textContent = `${draft.title} (${draft.author || 'Unknown'})`;
                select.appendChild(option);
            });

            dialogContainer.appendChild(select);

            const createBtn = document.createElement('button');
            createBtn.textContent = 'Create Preview';
            createBtn.className = 'btn btn-primary';
            createBtn.onclick = function() {
                const selectedSlug = select.value;
                if (selectedSlug) {
                    closeModal();
                    createPreviewSession(selectedSlug);
                } else {
                    showNotification('Error', 'Please select an article', 'warning');
                }
            };
            dialogContainer.appendChild(createBtn);
        } else {
            const noArticlesMsg = document.createElement('p');
            const em = document.createElement('em');
            em.textContent = 'No draft articles available for preview.';
            noArticlesMsg.appendChild(em);
            dialogContainer.appendChild(noArticlesMsg);
        }

        showModalWithElement(
            '👁️ Create Preview Session',
            dialogContainer,
            null,
            true
        );
    })
    .catch(error => {
        console.error('Failed to fetch drafts:', error);
        showNotification(
            'Error',
            'Failed to load draft articles',
            'error'
        );
    });
}

// Draft management functions
function refreshDraftList() {
    const tableBody = document.getElementById('drafts-table-body');
    if (!tableBody) return;

    // Show loading state
    tableBody.innerHTML = '<tr><td colspan="5" class="loading-row">Loading drafts...</td></tr>';

    fetch('/admin/drafts', {
        method: 'GET',
        headers: {
            'Accept': 'application/json'
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        populateDraftTable(data.drafts || []);
    })
    .catch(error => {
        console.error('Failed to load drafts:', error);
        tableBody.innerHTML = '<tr><td colspan="5" class="loading-row" style="color: #dc3545;">Failed to load drafts</td></tr>';
        showNotification('Error', 'Failed to load draft articles', 'error');
    });
}

function populateDraftTable(drafts) {
    const tableBody = document.getElementById('drafts-table-body');
    if (!tableBody) return;

    if (!drafts || drafts.length === 0) {
        tableBody.innerHTML = '<tr><td colspan="5" class="loading-row">No draft articles found</td></tr>';
        return;
    }

    tableBody.innerHTML = '';

    drafts.forEach(draft => {
        const row = document.createElement('tr');

        // Title cell
        const titleCell = document.createElement('td');
        titleCell.className = 'title-cell';
        titleCell.textContent = draft.title || 'Untitled';
        titleCell.title = draft.title || 'Untitled';
        row.appendChild(titleCell);

        // Date cell
        const dateCell = document.createElement('td');
        dateCell.className = 'date-cell';
        if (draft.date) {
            const date = new Date(draft.date);
            dateCell.textContent = date.toLocaleDateString();
        } else {
            dateCell.textContent = 'No date';
        }
        row.appendChild(dateCell);

        // Categories cell
        const categoriesCell = document.createElement('td');
        categoriesCell.className = 'categories-cell';
        if (draft.categories && draft.categories.length > 0) {
            categoriesCell.textContent = draft.categories.join(', ');
            categoriesCell.title = draft.categories.join(', ');
        } else {
            categoriesCell.textContent = 'None';
        }
        row.appendChild(categoriesCell);

        // Tags cell
        const tagsCell = document.createElement('td');
        tagsCell.className = 'categories-cell'; // Reuse same styling
        if (draft.tags && draft.tags.length > 0) {
            tagsCell.textContent = draft.tags.join(', ');
            tagsCell.title = draft.tags.join(', ');
        } else {
            tagsCell.textContent = 'None';
        }
        row.appendChild(tagsCell);

        // Actions cell
        const actionsCell = document.createElement('td');

        // Preview button
        const previewBtn = document.createElement('button');
        previewBtn.className = 'draft-action-btn draft-preview-btn';
        previewBtn.innerHTML = '👁️ Preview';
        previewBtn.title = 'Create preview session';
        previewBtn.onclick = () => createPreviewSession(draft.slug);
        actionsCell.appendChild(previewBtn);

        // View button
        const viewBtn = document.createElement('button');
        viewBtn.className = 'draft-action-btn draft-view-btn';
        viewBtn.innerHTML = '📖 View';
        viewBtn.title = 'View draft details';
        viewBtn.onclick = () => viewDraftDetails(draft.slug);
        actionsCell.appendChild(viewBtn);

        // Publish button
        const publishBtn = document.createElement('button');
        publishBtn.className = 'draft-action-btn draft-publish-btn';
        publishBtn.innerHTML = '🚀 Publish';
        publishBtn.title = 'Publish this draft';
        publishBtn.onclick = () => publishDraft(draft.slug);
        actionsCell.appendChild(publishBtn);

        row.appendChild(actionsCell);
        tableBody.appendChild(row);
    });
}

function viewDraftDetails(slug) {
    fetch(`/admin/drafts/${slug}`, {
        method: 'GET',
        headers: {
            'Accept': 'application/json'
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        return response.json();
    })
    .then(draft => {
        const detailsContainer = document.createElement('div');

        const title = document.createElement('h3');
        title.textContent = draft.title || 'Untitled';
        title.style.marginTop = '0';
        detailsContainer.appendChild(title);

        const meta = document.createElement('div');
        meta.style.fontSize = '0.9rem';
        meta.style.color = '#6b7280';
        meta.style.marginBottom = '15px';

        const metaInfo = [];
        if (draft.author) metaInfo.push(`Author: ${draft.author}`);
        if (draft.date) metaInfo.push(`Date: ${new Date(draft.date).toLocaleDateString()}`);
        if (draft.slug) metaInfo.push(`Slug: ${draft.slug}`);

        meta.textContent = metaInfo.join(' • ');
        detailsContainer.appendChild(meta);

        if (draft.description) {
            const desc = document.createElement('p');
            desc.textContent = draft.description;
            desc.style.fontStyle = 'italic';
            desc.style.marginBottom = '15px';
            detailsContainer.appendChild(desc);
        }

        if (draft.tags && draft.tags.length > 0) {
            const tagsDiv = document.createElement('div');
            tagsDiv.style.marginBottom = '10px';
            const tagsLabel = document.createElement('strong');
            tagsLabel.textContent = 'Tags: ';
            tagsDiv.appendChild(tagsLabel);
            tagsDiv.appendChild(document.createTextNode(draft.tags.join(', ')));
            detailsContainer.appendChild(tagsDiv);
        }

        if (draft.categories && draft.categories.length > 0) {
            const catsDiv = document.createElement('div');
            catsDiv.style.marginBottom = '15px';
            const catsLabel = document.createElement('strong');
            catsLabel.textContent = 'Categories: ';
            catsDiv.appendChild(catsLabel);
            catsDiv.appendChild(document.createTextNode(draft.categories.join(', ')));
            detailsContainer.appendChild(catsDiv);
        }

        // Action buttons
        const buttonsDiv = document.createElement('div');
        buttonsDiv.style.marginTop = '20px';

        const previewBtn = document.createElement('button');
        previewBtn.className = 'btn btn-primary';
        previewBtn.style.marginRight = '10px';
        previewBtn.textContent = '👁️ Create Preview';
        previewBtn.onclick = () => {
            closeModal();
            createPreviewSession(draft.slug);
        };
        buttonsDiv.appendChild(previewBtn);

        const publishBtn = document.createElement('button');
        publishBtn.className = 'btn btn-success';
        publishBtn.textContent = '🚀 Publish';
        publishBtn.onclick = () => {
            closeModal();
            publishDraft(draft.slug);
        };
        buttonsDiv.appendChild(publishBtn);

        detailsContainer.appendChild(buttonsDiv);

        showModalWithElement(
            `📝 Draft: ${draft.title}`,
            detailsContainer,
            null,
            true
        );
    })
    .catch(error => {
        console.error('Failed to load draft details:', error);
        showNotification('Error', 'Failed to load draft details', 'error');
    });
}

function publishDraft(slug) {
    if (!confirm(`Are you sure you want to publish the draft "${slug}"?`)) {
        return;
    }

    fetch(`/admin/drafts/${slug}/publish`, {
        method: 'POST',
        headers: {
            'Accept': 'application/json'
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        showNotification('Success', `Draft "${slug}" published successfully!`, 'success');
        // Refresh the draft list to remove the published article
        refreshDraftList();
    })
    .catch(error => {
        console.error('Failed to publish draft:', error);
        showNotification('Error', 'Failed to publish draft', 'error');
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

    // Add event listeners for preview buttons (delegated event handling)
    document.addEventListener('click', function(event) {
        if (event.target.classList.contains('create-preview-btn')) {
            showCreatePreviewDialog();
        }
    });

    // Load draft list on page load if drafts table exists
    setTimeout(() => {
        const draftsTable = document.getElementById('drafts-table');
        if (draftsTable) {
            refreshDraftList();
        }
    }, 100);
});