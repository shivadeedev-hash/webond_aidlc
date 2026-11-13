import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient, filenameFromDisposition, setAccessToken } from '../lib/apiClient';
import { ApiClientError } from '../lib/errors';

/** jsdom's Blob has no .text(); read it through FileReader instead. */
function blobText(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result));
    reader.onerror = () => reject(reader.error);
    reader.readAsText(blob);
  });
}

function fileResponse(body: string, headers: Record<string, string> = {}, status = 200): Response {
  return new Response(body, { status, headers });
}

function jsonResponse(body: unknown, status: number): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('apiClient.download', () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    vi.stubGlobal('fetch', fetchMock);
    setAccessToken('tok-123');
  });

  afterEach(() => {
    fetchMock.mockReset();
    vi.unstubAllGlobals();
    setAccessToken(null);
  });

  it('returns the blob and the file name from Content-Disposition, sending the bearer token', async () => {
    fetchMock.mockResolvedValueOnce(
      fileResponse('PK-binary', {
        'Content-Type': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'Content-Disposition': 'attachment; filename="webond-transactions-20261028.xlsx"',
      }),
    );
    const file = await apiClient.download('/transactions/export', {
      query: { tab: 'all', q: undefined },
      fallbackFilename: 'fallback.xlsx',
    });
    expect(file.filename).toBe('webond-transactions-20261028.xlsx');
    expect(await blobText(file.blob)).toBe('PK-binary');

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toContain('/transactions/export?tab=all');
    expect(url).not.toContain('q=');
    expect((init.headers as Record<string, string>).Authorization).toBe('Bearer tok-123');
    expect(init.credentials).toBe('include');
  });

  it('falls back to the supplied file name when the header is not exposed', async () => {
    fetchMock.mockResolvedValueOnce(
      fileResponse('%PDF-1.3', { 'Content-Type': 'application/pdf' }),
    );
    const file = await apiClient.download('/statements/202610', {
      fallbackFilename: 'webond-statement-202610.pdf',
    });
    expect(file.filename).toBe('webond-statement-202610.pdf');
  });

  it('turns an error envelope into a typed ApiClientError', async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse(
        {
          error: { code: 'EXPORT_TOO_LARGE', message: 'Export exceeds 10,000 rows' },
          meta: { requestId: 'r-1' },
        },
        400,
      ),
    );
    await expect(
      apiClient.download('/transactions/export', { fallbackFilename: 'x.xlsx' }),
    ).rejects.toMatchObject({ code: 'EXPORT_TOO_LARGE', status: 400 });
  });

  it('reports a network failure as NETWORK_ERROR', async () => {
    fetchMock.mockRejectedValueOnce(new TypeError('failed to fetch'));
    await expect(
      apiClient.download('/statements/202610', { fallbackFilename: 'x.pdf' }),
    ).rejects.toBeInstanceOf(ApiClientError);
  });

  it('refreshes the token once on 401 and replays the download', async () => {
    fetchMock
      .mockResolvedValueOnce(
        jsonResponse({ error: { code: 'UNAUTHORIZED', message: 'expired' } }, 401),
      )
      .mockResolvedValueOnce(jsonResponse({ data: { accessToken: 'tok-new' } }, 200)) // POST /auth/refresh
      .mockResolvedValueOnce(fileResponse('%PDF-1.3', { 'Content-Type': 'application/pdf' }));
    const file = await apiClient.download('/statements/202610', { fallbackFilename: 'x.pdf' });
    expect(await blobText(file.blob)).toBe('%PDF-1.3');
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(String(fetchMock.mock.calls[1][0])).toContain('/auth/refresh');
    const replayInit = fetchMock.mock.calls[2][1] as RequestInit;
    expect((replayInit.headers as Record<string, string>).Authorization).toBe('Bearer tok-new');
  });
});

describe('401 on /auth/* does not trigger a refresh attempt', () => {
  const fetchMock = vi.fn();
  beforeEach(() => vi.stubGlobal('fetch', fetchMock));
  afterEach(() => {
    fetchMock.mockReset();
    vi.unstubAllGlobals();
  });

  it('a wrong password is reported directly (one request, no refresh)', async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse({ error: { code: 'UNAUTHORIZED', message: 'Invalid email or password' } }, 401),
    );
    await expect(
      apiClient.post('/auth/login', { email: 'a', password: 'b' }),
    ).rejects.toMatchObject({
      code: 'UNAUTHORIZED',
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});

describe('filenameFromDisposition', () => {
  it('reads quoted, unquoted and RFC 5987 forms', () => {
    expect(filenameFromDisposition('attachment; filename="a.pdf"')).toBe('a.pdf');
    expect(filenameFromDisposition('attachment; filename=b.xlsx')).toBe('b.xlsx');
    expect(filenameFromDisposition("attachment; filename*=UTF-8''c%20d.pdf")).toBe('c d.pdf');
    expect(filenameFromDisposition(null)).toBeNull();
    expect(filenameFromDisposition('inline')).toBeNull();
  });
});
