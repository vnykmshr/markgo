/**
 * Authenticated Fetch — single source of truth for mutation fetches.
 *
 * Handles CSRF token injection, auth error detection (401/403), and
 * JSON response parsing. All POST/PUT/DELETE requests go through this
 * module. GET-only navigation (router.js) stays separate.
 */

/**
 * Read CSRF token from the <meta name="csrf-token"> tag.
 * @returns {string} The token value, or empty string if not found.
 */
export function getCSRFToken() {
    return document.querySelector('meta[name="csrf-token"]')?.content || '';
}

/**
 * Fetch wrapper that injects CSRF token and detects auth errors.
 *
 * - Adds X-CSRF-Token header (unless skipCSRF is set)
 * - Dispatches auth:open-login on 401/403
 * - Throws Error if CSRF token is missing (no network request made)
 * - Throws TypeError on network failure (from native fetch)
 *
 * @param {string} url
 * @param {RequestInit & { skipCSRF?: boolean }} options
 * @returns {Promise<Response>}
 */
export async function authenticatedFetch(url, options = {}) {
    const { skipCSRF, ...fetchOptions } = options;

    if (!skipCSRF) {
        const token = getCSRFToken();
        if (!token) {
            throw new Error('Session expired \u2014 please refresh the page');
        }
        fetchOptions.headers = { ...fetchOptions.headers, 'X-CSRF-Token': token };
    }

    const response = await fetch(url, fetchOptions);

    if (response.status === 401 || response.status === 403) {
        document.dispatchEvent(new CustomEvent('auth:open-login'));
    }

    return response;
}

/**
 * Fetch JSON with CSRF and auth handling. Returns a structured result
 * instead of throwing on HTTP errors.
 *
 * - Auto-stringifies object bodies
 * - Sets Content-Type and Accept headers for JSON
 * - Returns { ok, status, data, error } on all code paths
 * - Throws TypeError on network failure (caller must handle)
 *
 * @param {string} url
 * @param {RequestInit & { skipCSRF?: boolean }} options
 * @returns {Promise<{ ok: boolean, status: number, data: any, error: string | null }>}
 */
export async function authenticatedJSON(url, options = {}) {
    let body = options.body;
    if (body !== undefined && typeof body !== 'string') {
        body = JSON.stringify(body);
    }

    let response;
    try {
        response = await authenticatedFetch(url, {
            ...options,
            body,
            headers: {
                'Content-Type': 'application/json',
                Accept: 'application/json',
                ...options.headers,
            },
        });
    } catch (err) {
        if (err instanceof TypeError) throw err;
        return { ok: false, status: 0, data: null, error: err.message };
    }

    if (response.status === 401 || response.status === 403) {
        return { ok: false, status: response.status, data: null, error: 'Please sign in to continue' };
    }

    let data;
    try {
        data = await response.json();
    } catch {
        return { ok: false, status: response.status, data: null, error: `Server error (${response.status})` };
    }

    return {
        ok: response.ok,
        status: response.status,
        data,
        error: response.ok ? null : (data.error || `Request failed (${response.status})`),
    };
}
