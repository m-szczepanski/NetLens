import { PUBLIC_BACKEND_HTTP_URL, PUBLIC_BACKEND_WS_URL } from '$env/static/public';
import { browser } from '$app/environment';

/**
 * Default development backend: the Go server running locally on its own port,
 * separate from the Vite dev server. In production the UI is served by the
 * NetLens binary itself, so the backend lives on the same origin.
 */
export const DEV_BACKEND_DEFAULT = 'http://localhost:8080';

export interface BackendConfig {
	/** Absolute base URL of the REST API, e.g. http://localhost:8080. */
	httpBaseUrl: string;
	/** Absolute base URL of the WebSocket API, e.g. ws://localhost:8080. */
	wsBaseUrl: string;
}

/** Derive a ws(s):// base URL from an http(s):// base URL. */
export function toWebSocketUrl(httpBaseUrl: string): string {
	return httpBaseUrl.replace(/^http/, 'ws');
}

/**
 * Resolve the backend base URLs. Empty overrides fall back to:
 * the current origin in the browser, DEV_BACKEND_DEFAULT otherwise.
 */
export function resolveBackendConfig(
	httpOverride: string,
	wsOverride: string,
	currentOrigin: string
): BackendConfig {
	const httpBaseUrl = httpOverride || currentOrigin;
	return { httpBaseUrl, wsBaseUrl: wsOverride || toWebSocketUrl(httpBaseUrl) };
}

export const backend: BackendConfig = resolveBackendConfig(
	PUBLIC_BACKEND_HTTP_URL,
	PUBLIC_BACKEND_WS_URL,
	browser ? window.location.origin : DEV_BACKEND_DEFAULT
);
