import { describe, expect, it } from 'vitest';
import { DEV_BACKEND_DEFAULT, resolveBackendConfig, toWebSocketUrl } from './config';

describe('toWebSocketUrl', () => {
	it('derives ws:// from http://', () => {
		expect(toWebSocketUrl('http://localhost:8080')).toBe('ws://localhost:8080');
	});

	it('derives wss:// from https://', () => {
		expect(toWebSocketUrl('https://netlens.example')).toBe('wss://netlens.example');
	});
});

describe('resolveBackendConfig', () => {
	it('falls back to the dev default outside the browser', () => {
		expect(resolveBackendConfig('', '', DEV_BACKEND_DEFAULT)).toEqual({
			httpBaseUrl: 'http://localhost:8080',
			wsBaseUrl: 'ws://localhost:8080'
		});
	});

	it('falls back to the current origin when embedded in the backend', () => {
		expect(resolveBackendConfig('', '', 'http://192.168.1.50:8080').httpBaseUrl).toBe(
			'http://192.168.1.50:8080'
		);
	});

	it('honours explicit environment overrides', () => {
		expect(resolveBackendConfig('https://api.example', 'wss://ws.example', '')).toEqual({
			httpBaseUrl: 'https://api.example',
			wsBaseUrl: 'wss://ws.example'
		});
	});
});
