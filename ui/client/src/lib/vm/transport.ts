import type { BaseQueryFn } from '@reduxjs/toolkit/query/react';
import { API_BASE_URL } from '../types';

const ABSOLUTE_HTTP_URL_RE = /^https?:\/\//i;

export interface ApiRequestArgs {
  url: string;
  method?: string;
  body?: unknown;
  query?: Record<string, string | number | undefined>;
}

export interface ApiError {
  status: number;
  message: string;
}

function getURL(path: string, query?: Record<string, string | number | undefined>): string {
  const hasAbsoluteBase = ABSOLUTE_HTTP_URL_RE.test(API_BASE_URL);
  const baseURL = hasAbsoluteBase
    ? new URL(API_BASE_URL.endsWith('/') ? API_BASE_URL : `${API_BASE_URL}/`)
    : new URL(window.location.origin);
  const url = hasAbsoluteBase
    ? new URL(path, baseURL)
    : new URL(`${API_BASE_URL}${path}`, baseURL);

  if (query) {
    Object.entries(query).forEach(([key, value]) => {
      if (value === undefined || value === '') return;
      url.searchParams.set(key, String(value));
    });
  }

  return hasAbsoluteBase ? url.toString() : `${url.pathname}${url.search}`;
}

export type VmBaseQuery = BaseQueryFn<ApiRequestArgs | string, unknown, ApiError>;

export const vmBaseQuery: VmBaseQuery = async (args) => {
  const { url, method = 'GET', body, query } =
    typeof args === 'string' ? { url: args } : args;

  try {
    const response = await fetch(getURL(url, query), {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });

    if (!response.ok) {
      let message = `HTTP ${response.status}`;
      try {
        const envelope = await response.json();
        if (envelope?.error?.message) {
          message = String(envelope.error.message);
        }
      } catch {
        // ignore response body parse failures for error envelopes
      }
      return { error: { status: response.status, message } };
    }

    if (response.status === 204) {
      return { data: undefined };
    }

    const data = await response.json();
    return { data };
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : 'Network error';
    return { error: { status: 0, message } };
  }
};
