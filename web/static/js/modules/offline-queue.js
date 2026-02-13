/**
 * Offline Compose Queue — IndexedDB-backed queue for offline posts.
 *
 * When publish fails due to network error, the post is queued.
 * When connectivity returns, the queue is drained automatically.
 * CSRF token is read fresh from <meta> on each drain attempt.
 */

const DB_NAME = 'markgo';
const STORE_NAME = 'compose-queue';
const DB_VERSION = 1;

function openDB() {
    return new Promise((resolve, reject) => {
        const request = indexedDB.open(DB_NAME, DB_VERSION);
        request.onupgradeneeded = () => {
            const db = request.result;
            if (!db.objectStoreNames.contains(STORE_NAME)) {
                db.createObjectStore(STORE_NAME, { keyPath: 'id', autoIncrement: true });
            }
        };
        request.onsuccess = () => resolve(request.result);
        request.onerror = () => reject(request.error);
    });
}

/**
 * Add a compose input to the offline queue.
 * @param {{ content: string, title?: string }} input
 */
export async function queuePost(input) {
    const db = await openDB();
    return new Promise((resolve, reject) => {
        const tx = db.transaction(STORE_NAME, 'readwrite');
        tx.objectStore(STORE_NAME).add({ ...input, queuedAt: Date.now() });
        tx.oncomplete = () => { db.close(); resolve(); };
        tx.onerror = () => { db.close(); reject(tx.error); };
    });
}

/**
 * Get number of queued posts.
 */
export async function getQueueCount() {
    const db = await openDB();
    return new Promise((resolve, reject) => {
        const tx = db.transaction(STORE_NAME, 'readonly');
        const req = tx.objectStore(STORE_NAME).count();
        req.onsuccess = () => { db.close(); resolve(req.result); };
        req.onerror = () => { db.close(); reject(req.error); };
    });
}

/**
 * Drain the queue: POST each item to /compose/quick.
 * Returns { published: number, failed: number }.
 * failed === -1 signals missing CSRF token (caller should warn user).
 */
export async function drainQueue() {
    const token = document.querySelector('meta[name="csrf-token"]')?.content;
    if (!token) {
        console.warn('drainQueue: no CSRF token available, cannot sync');
        return { published: 0, failed: -1 };
    }

    let db;
    try {
        db = await openDB();
    } catch (err) {
        console.warn('drainQueue: IndexedDB unavailable:', err?.message || err);
        return { published: 0, failed: 0 };
    }

    try {
        const items = await new Promise((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, 'readonly');
            const req = tx.objectStore(STORE_NAME).getAll();
            req.onsuccess = () => resolve(req.result);
            req.onerror = () => reject(req.error);
        });

        if (items.length === 0) return { published: 0, failed: 0 };

        let published = 0;
        let failed = 0;

        for (const item of items) {
            const body = { content: item.content };
            if (item.title) body.title = item.title;

            try {
                const response = await fetch('/compose/quick', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': token,
                    },
                    body: JSON.stringify(body),
                });

                if (response.ok) {
                    // Remove from queue on success
                    await new Promise((resolve, reject) => {
                        const tx = db.transaction(STORE_NAME, 'readwrite');
                        tx.objectStore(STORE_NAME).delete(item.id);
                        tx.oncomplete = () => resolve();
                        tx.onerror = () => reject(tx.error);
                    });
                    published++;
                } else if (response.status === 401 || response.status === 403) {
                    // Auth/CSRF expired — stop draining, keep items for retry after login
                    failed += items.length - published;
                    break;
                } else {
                    failed++;
                }
            } catch (err) {
                // Network error or unexpected failure — stop draining
                console.warn('Queue drain stopped:', err?.message || err);
                failed += items.length - published;
                break;
            }
        }

        return { published, failed };
    } finally {
        if (db) db.close();
    }
}
