import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { AuthProvider } from '../hooks/AuthProvider';
import { useAuth } from '../hooks/useAuth';
import { I18nProvider } from '../i18n/I18nProvider';
import ProtectedRoute from '../components/layout/ProtectedRoute';
import { SESSION_HINT_KEY } from '../lib/constants';
import { ApiClientError } from '../lib/errors';
import { getAccessToken, setAccessToken } from '../lib/apiClient';
import type { User } from '../types';

const { refresh } = vi.hoisted(() => ({ refresh: vi.fn() }));
vi.mock('../services/auth', () => ({
  authService: { refresh, login: vi.fn(), logout: vi.fn().mockResolvedValue(undefined) },
}));

const USER = { id: 7, email: 'restored@webond.test', fullName: 'Restored User' } as User;

function Who(): JSX.Element {
  const { user, isInitializing } = useAuth();
  return <p>{isInitializing ? 'initializing' : (user?.email ?? 'guest')}</p>;
}

function renderTree(restoreSession: boolean, ui: JSX.Element = <Who />): void {
  render(
    <I18nProvider initialLocale="en">
      <AuthProvider restoreSession={restoreSession}>
        <MemoryRouter initialEntries={['/private']}>{ui}</MemoryRouter>
      </AuthProvider>
    </I18nProvider>,
  );
}

describe('session restore after a page reload (D3-4)', () => {
  beforeEach(() => {
    refresh.mockReset();
    window.localStorage.clear();
    setAccessToken(null);
  });

  it('restores the user and token from the refresh cookie when a session hint exists', async () => {
    window.localStorage.setItem(SESSION_HINT_KEY, '1');
    refresh.mockResolvedValueOnce({ accessToken: 'fresh-token', user: USER });

    renderTree(true);

    expect(screen.getByText('initializing')).toBeInTheDocument();
    expect(await screen.findByText('restored@webond.test')).toBeInTheDocument();
    expect(getAccessToken()).toBe('fresh-token');
  });

  it('does not call the API for a first-time visitor (no hint, no pointless 401)', async () => {
    renderTree(true);

    expect(await screen.findByText('guest')).toBeInTheDocument();
    expect(refresh).not.toHaveBeenCalled();
  });

  it('falls back to guest and forgets the hint when the session has expired', async () => {
    window.localStorage.setItem(SESSION_HINT_KEY, '1');
    refresh.mockRejectedValueOnce(new ApiClientError('UNAUTHORIZED'));

    renderTree(true);

    expect(await screen.findByText('guest')).toBeInTheDocument();
    expect(window.localStorage.getItem(SESSION_HINT_KEY)).toBeNull();
    expect(getAccessToken()).toBeNull();
  });

  it('never restores in mock mode', async () => {
    window.localStorage.setItem(SESSION_HINT_KEY, '1');
    renderTree(false);

    expect(await screen.findByText('guest')).toBeInTheDocument();
    expect(refresh).not.toHaveBeenCalled();
  });

  it('ProtectedRoute waits for the restore instead of bouncing to /login', async () => {
    window.localStorage.setItem(SESSION_HINT_KEY, '1');
    let resolveRefresh: (v: { accessToken: string; user: User }) => void = () => {};
    refresh.mockReturnValueOnce(new Promise((resolve) => (resolveRefresh = resolve)));

    renderTree(
      true,
      <Routes>
        <Route element={<ProtectedRoute />}>
          <Route path="/private" element={<p>private page</p>} />
        </Route>
        <Route path="/login" element={<p>login page</p>} />
      </Routes>,
    );

    // While the restore is pending we must not have redirected.
    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.queryByText('login page')).not.toBeInTheDocument();

    resolveRefresh({ accessToken: 't', user: USER });
    await waitFor(() => expect(screen.getByText('private page')).toBeInTheDocument());
  });

  it('ProtectedRoute sends a guest to /login once the restore fails', async () => {
    window.localStorage.setItem(SESSION_HINT_KEY, '1');
    refresh.mockRejectedValueOnce(new ApiClientError('UNAUTHORIZED'));

    renderTree(
      true,
      <Routes>
        <Route element={<ProtectedRoute />}>
          <Route path="/private" element={<p>private page</p>} />
        </Route>
        <Route path="/login" element={<p>login page</p>} />
      </Routes>,
    );

    expect(await screen.findByText('login page')).toBeInTheDocument();
  });
});
