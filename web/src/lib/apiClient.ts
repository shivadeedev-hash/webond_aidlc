// Thin fetch wrapper for the WeBond REST API (api-spec.md).
// - Reads/writes the access token from an in-memory store (never localStorage).
// - Unwraps the { data, meta } / { error, meta } envelope.
// - Throws ApiClientError (typed code) on error responses.
// - Silent refresh: on a 401 it attempts POST /auth/refresh once (the refresh
//   token lives in an httpOnly cookie) and replays the request. Requests to
//   /auth/* never trigger it (a failed login or refresh is not an expired session).
// - download(): authenticated binary downloads (xlsx / pdf) with the same
//   auth + refresh behaviour.

import { API_URL } from './constants';
import { ApiClientError, type ApiErrorCode } from './errors';
import type { ApiEnvelope, DownloadedFile, Meta } from '../types/common';

// ---- In-memory access-token store (shared with useAuth) ----
let accessToken: string | null = null;

export function setAccessToken(token: string | null): void {
  accessToken = token;
}

export function getAccessToken(): string | null {
  return accessToken;
}

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  /** JSON body; serialized automatically. */
  body?: unknown;
  /** Extra query params appended to the path. */
  query?: Record<string, string | number | boolean | undefined | null>;
  /** Extra headers (e.g. Idempotency-Key). */
  headers?: Record<string, string>;
  /** Internal: prevents infinite refresh recursion. */
  _retry?: boolean;
}

/** Result of a request: the unwrapped payload plus envelope meta. */
export interface ApiResult<T> {
  data: T;
  meta?: Meta;
}

function buildUrl(path: string, query?: RequestOptions['query']): string {
  const base = `${API_URL}${path.startsWith('/') ? path : `/${path}`}`;
  if (!query) return base;
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined && value !== null) params.append(key, String(value));
  }
  const qs = params.toString();
  return qs ? `${base}?${qs}` : base;
}

function isErrorEnvelope<T>(body: ApiEnvelope<T>): body is {
  error: { code: string; message: string; details?: Record<string, unknown> };
  meta?: Meta;
} {
  return typeof body === 'object' && body !== null && 'error' in body;
}

/** Silent-refresh stub: rotate the access token via the refresh cookie (EX-03). */
async function tryRefresh(): Promise<boolean> {
  try {
    const res = await fetch(buildUrl('/auth/refresh'), {
      method: 'POST',
      credentials: 'include',
    });
    if (!res.ok) return false;
    const body = (await res.json()) as ApiEnvelope<{ accessToken: string }>;
    if (isErrorEnvelope(body)) return false;
    setAccessToken(body.data.accessToken);
    return true;
  } catch {
    return false;
  }
}

export async function request<T>(
  path: string,
  options: RequestOptions = {},
): Promise<ApiResult<T>> {
  const { method = 'GET', body, query, headers = {}, _retry = false } = options;

  const finalHeaders: Record<string, string> = { ...headers };
  if (body !== undefined) finalHeaders['Content-Type'] = 'application/json';
  if (accessToken) finalHeaders['Authorization'] = `Bearer ${accessToken}`;

  let res: Response;
  try {
    res = await fetch(buildUrl(path, query), {
      method,
      headers: finalHeaders,
      credentials: 'include',
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch {
    throw new ApiClientError('NETWORK_ERROR');
  }

  // 401 → attempt a single silent refresh, then replay the original request.
  // Not for /auth/* (wrong password, or the refresh call itself failing).
  if (res.status === 401 && !_retry && !path.startsWith('/auth/')) {
    const refreshed = await tryRefresh();
    if (refreshed) {
      return request<T>(path, { ...options, _retry: true });
    }
  }

  // 204 No Content.
  if (res.status === 204) {
    return { data: undefined as unknown as T };
  }

  let parsed: ApiEnvelope<T>;
  try {
    parsed = (await res.json()) as ApiEnvelope<T>;
  } catch {
    if (!res.ok) {
      throw new ApiClientError('UNKNOWN', undefined, { status: res.status });
    }
    return { data: undefined as unknown as T };
  }

  if (!res.ok || isErrorEnvelope(parsed)) {
    const err = isErrorEnvelope(parsed) ? parsed.error : undefined;
    throw new ApiClientError(err?.code as ApiErrorCode | undefined, err?.message, {
      status: res.status,
      details: err?.details,
      requestId: parsed.meta?.requestId,
    });
  }

  return { data: parsed.data, meta: parsed.meta };
}

export interface DownloadOptions {
  query?: RequestOptions['query'];
  /** File name used when the response carries no Content-Disposition header. */
  fallbackFilename: string;
  /** Internal: prevents infinite refresh recursion. */
  _retry?: boolean;
}

/** Extracts the file name from a Content-Disposition header, if present. */
export function filenameFromDisposition(header: string | null): string | null {
  if (!header) return null;
  const utf8 = /filename\*=UTF-8''([^;]+)/i.exec(header);
  if (utf8) {
    try {
      return decodeURIComponent(utf8[1].trim());
    } catch {
      /* fall through to the plain filename */
    }
  }
  const plain = /filename="?([^";]+)"?/i.exec(header);
  return plain ? plain[1].trim() : null;
}

/**
 * Authenticated GET of a binary file. Error responses (JSON envelopes such as
 * EXPORT_TOO_LARGE) are turned into a typed ApiClientError, exactly like request().
 */
export async function download(path: string, options: DownloadOptions): Promise<DownloadedFile> {
  const { query, fallbackFilename, _retry = false } = options;
  const headers: Record<string, string> = {};
  if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;

  let res: Response;
  try {
    res = await fetch(buildUrl(path, query), { method: 'GET', headers, credentials: 'include' });
  } catch {
    throw new ApiClientError('NETWORK_ERROR');
  }

  if (res.status === 401 && !_retry && !path.startsWith('/auth/')) {
    if (await tryRefresh()) return download(path, { ...options, _retry: true });
  }

  const contentType = res.headers.get('Content-Type') ?? '';
  if (!res.ok || contentType.includes('application/json')) {
    let parsed: ApiEnvelope<unknown>;
    try {
      parsed = (await res.json()) as ApiEnvelope<unknown>;
    } catch {
      throw new ApiClientError('UNKNOWN', undefined, { status: res.status });
    }
    const err = isErrorEnvelope(parsed) ? parsed.error : undefined;
    throw new ApiClientError(err?.code as ApiErrorCode | undefined, err?.message, {
      status: res.status,
      details: err?.details,
      requestId: parsed.meta?.requestId,
    });
  }

  const blob = await res.blob();
  return {
    blob,
    filename: filenameFromDisposition(res.headers.get('Content-Disposition')) ?? fallbackFilename,
  };
}

export const apiClient = {
  get: <T>(path: string, query?: RequestOptions['query']) =>
    request<T>(path, { method: 'GET', query }),
  post: <T>(path: string, body?: unknown, extra?: Omit<RequestOptions, 'method' | 'body'>) =>
    request<T>(path, { method: 'POST', body, ...extra }),
  put: <T>(path: string, body?: unknown) => request<T>(path, { method: 'PUT', body }),
  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
  download,
};
