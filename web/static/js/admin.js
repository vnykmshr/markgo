/**
 * Admin Dashboard — notifications, modals, GET/POST action execution.
 */

let ac = null;

export function init() {
    ac = new AbortController();
    const { signal } = ac;
    const container = document.getElementById('notification-container');

    // =========================================================================
    // Notifications
    // =========================================================================

    function showNotification(title, message, type = 'success') {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;

        const closeBtn = document.createElement('button');
        closeBtn.className = 'notification-close';
        closeBtn.textContent = '\u00d7';
        notification.appendChild(closeBtn);

        const titleEl = document.createElement('div');
        titleEl.className = 'notification-title';
        titleEl.textContent = title;
        notification.appendChild(titleEl);

        const messageEl = document.createElement('div');
        messageEl.className = 'notification-message';
        messageEl.textContent = message;
        notification.appendChild(messageEl);

        container.appendChild(notification);

        closeBtn.addEventListener('click', () => removeNotification(notification));
        setTimeout(() => {
            if (notification.parentNode) removeNotification(notification);
        }, 5000);
    }

    function removeNotification(notification) {
        notification.classList.add('notification-slide-out');
        setTimeout(() => {
            if (notification.parentNode) notification.parentNode.removeChild(notification);
        }, 300);
    }

    // =========================================================================
    // Modal
    // =========================================================================

    const modal = document.getElementById('result-modal');

    function showModalWithElement(title, messageElement, data = null, isSuccess = true) {
        document.getElementById('modal-title').textContent = title;

        const messageContainer = document.getElementById('modal-message');
        messageContainer.textContent = '';
        messageContainer.appendChild(messageElement);

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

        const header = document.querySelector('.modal-header');
        header.classList.remove('modal-header-success', 'modal-header-error', 'modal-header-default');
        header.classList.add(isSuccess ? 'modal-header-success' : 'modal-header-error');

        modal.className = 'modal modal-visible';
        document.body.classList.add('body-modal-open');
        document.body.classList.remove('body-modal-closed');
    }

    function closeModal() {
        modal.className = 'modal modal-hidden';
        document.body.classList.add('body-modal-closed');
        document.body.classList.remove('body-modal-open');
    }

    // =========================================================================
    // Result message builders
    // =========================================================================

    function buildTimestamp() {
        const p = document.createElement('p');
        const small = document.createElement('small');
        small.textContent = `Timestamp: ${new Date().toLocaleString()}`;
        p.appendChild(small);
        return p;
    }

    function buildSuccessMessage(data, actionName) {
        const el = document.createElement('div');

        const msg = document.createElement('p');
        const strong = document.createElement('strong');
        strong.textContent = 'Data retrieved successfully';
        msg.appendChild(strong);
        el.appendChild(msg);

        if (data.articles) {
            const stats = document.createElement('p');
            stats.textContent = `Total: ${data.articles.total || 0} articles (${data.articles.published || 0} published, ${data.articles.drafts || 0} drafts)`;
            el.appendChild(stats);
        } else if (data.status === 'healthy') {
            const status = document.createElement('p');
            status.textContent = 'System Status: ';
            const span = document.createElement('span');
            span.className = 'status-healthy';
            span.textContent = 'Healthy';
            status.appendChild(span);
            el.appendChild(status);
            if (data.uptime) {
                const up = document.createElement('p');
                up.textContent = `Uptime: ${data.uptime}`;
                el.appendChild(up);
            }
        } else if (data.memory) {
            const mem = document.createElement('p');
            mem.textContent = `Memory Allocated: ${data.memory.alloc || 'N/A'}`;
            el.appendChild(mem);
            if (data.uptime) {
                const up = document.createElement('p');
                up.textContent = `Uptime: ${data.uptime}`;
                el.appendChild(up);
            }
        }

        el.appendChild(buildTimestamp());
        return el;
    }

    function buildErrorMessage(error) {
        const el = document.createElement('div');

        const title = document.createElement('p');
        const strong = document.createElement('strong');
        strong.textContent = 'Failed to load data:';
        title.appendChild(strong);
        el.appendChild(title);

        const msg = document.createElement('p');
        msg.textContent = error.message || 'Unknown error';
        el.appendChild(msg);

        el.appendChild(buildTimestamp());
        return el;
    }

    // =========================================================================
    // GET action execution
    // =========================================================================

    function executeGetAction(url, actionName) {
        const button = document.querySelector(`.admin-get-btn[data-url="${url}"][data-name="${actionName}"]`);
        if (!button) return;

        const originalText = button.textContent;
        button.textContent = 'Loading...';
        button.disabled = true;

        fetch(url, { method: 'GET', headers: { Accept: 'application/json' } })
            .then((response) => {
                if (!response.ok) throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                return response.json();
            })
            .then((data) => {
                showModalWithElement(`${actionName}`, buildSuccessMessage(data, actionName), data, true);
                showNotification('Data Loaded', `${actionName} data loaded successfully`, 'success');
            })
            .catch((error) => {
                console.error('Get action failed:', error);
                showModalWithElement(`${actionName} - Failed`, buildErrorMessage(error), error.stack || error.toString(), false);
                showNotification('Load Failed', error.message || `Failed to load ${actionName}`, 'error');
            })
            .finally(() => {
                button.textContent = originalText;
                button.disabled = false;
            });
    }

    // =========================================================================
    // POST action execution
    // =========================================================================

    function executeAdminAction(url, actionName) {
        if (!confirm(`Are you sure you want to execute "${actionName}"?`)) return;

        const button = document.querySelector(`.admin-action-btn[data-url="${url}"][data-name="${actionName}"]`);
        if (!button) return;

        const originalText = button.textContent;
        button.textContent = 'Executing...';
        button.disabled = true;

        fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json', Accept: 'application/json' } })
            .then((response) => {
                if (!response.ok) throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                return response.json();
            })
            .then((data) => {
                const el = document.createElement('div');
                const msg = document.createElement('p');
                const strong = document.createElement('strong');
                strong.textContent = data.message || `${actionName} completed successfully`;
                msg.appendChild(strong);
                el.appendChild(msg);
                el.appendChild(buildTimestamp());

                showModalWithElement(`${actionName} - Success`, el, data, true);
                showNotification('Action Completed', data.message || `${actionName} completed successfully`, 'success');

                if (url.includes('reload') || url.includes('cache/clear')) {
                    setTimeout(() => {
                        if (confirm('Action completed successfully. Refresh the page to see updated data?')) {
                            window.location.reload();
                        }
                    }, 2000);
                }
            })
            .catch((error) => {
                console.error('Admin action failed:', error);

                const el = document.createElement('div');
                const title = document.createElement('p');
                const strong = document.createElement('strong');
                strong.textContent = 'Action failed:';
                title.appendChild(strong);
                el.appendChild(title);
                const msg = document.createElement('p');
                msg.textContent = error.message || `Failed to execute ${actionName}`;
                el.appendChild(msg);
                el.appendChild(buildTimestamp());

                showModalWithElement(`${actionName} - Failed`, el, error.stack || error.toString(), false);
                showNotification('Action Failed', error.message || `Failed to execute ${actionName}`, 'error');
            })
            .finally(() => {
                button.textContent = originalText;
                button.disabled = false;
            });
    }

    // =========================================================================
    // Event binding
    // =========================================================================

    document.querySelectorAll('.admin-action-btn').forEach((button) => {
        button.addEventListener('click', () => executeAdminAction(button.dataset.url, button.dataset.name), { signal });
    });

    document.querySelectorAll('.admin-get-btn').forEach((button) => {
        button.addEventListener('click', () => executeGetAction(button.dataset.url, button.dataset.name), { signal });
    });

    document.querySelectorAll('#modal-close-btn, #modal-close-footer').forEach((btn) => {
        btn.addEventListener('click', closeModal, { signal });
    });

    window.addEventListener('click', (e) => {
        if (e.target === modal) closeModal();
    }, { signal });

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') closeModal();
    }, { signal });
}

export function destroy() {
    if (ac) { ac.abort(); ac = null; }
}
