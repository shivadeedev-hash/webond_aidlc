// Auth provider: holds the authenticated user and the access token (in memory,
// via apiClient — never persisted to localStorage, NFR-01). Guests have user === null.
//
// Session restore (D3-4 silent refresh): the refresh token lives in an httpOnly
// cookie, so after a page reload the provider asks POST /auth/refresh for a new
// access token and the user. To avoid a pointless 401 for every first-time
// visitor, it only tries when a non-secret "had a session" hint is present
// (set at login, cleared at logout or when the restore fails).
import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';
import { authService } from '../services/auth';
import { setAccessToken } from '../lib/apiClient';
import { DATA_SOURCE, SESSION_HINT_KEY } from '../lib/constants';
import type { LoginRequest, User } from '../types';
import { AuthContext, type AuthContextValue } from './authContext';

function readHint(): boolean {
  try {
    return window.localStorage.getItem(SESSION_HINT_KEY) === '1';
  } catch {
    return false;
  }
}

function writeHint(on: boolean): void {
  try {
    if (on) window.localStorage.setItem(SESSION_HINT_KEY, '1');
    else window.localStorage.removeItem(SESSION_HINT_KEY);
  } catch {
    /* storage unavailable (private mode): the session just won't survive a reload */
  }
}

export function AuthProvider({
  children,
  restoreSession = DATA_SOURCE === 'api',
}: {
  children: ReactNode;
  /** Restore a saved session on mount. Defaults to on when talking to the real API. */
  restoreSession?: boolean;
}): JSX.Element {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isInitializing, setIsInitializing] = useState(() => restoreSession && readHint());

  useEffect(() => {
    if (!restoreSession || !readHint()) {
      setIsInitializing(false);
      return;
    }
    let cancelled = false;
    authService
      .refresh()
      .then((res) => {
        if (cancelled) return;
        setAccessToken(res.accessToken);
        if (res.user) setUser(res.user);
        else writeHint(false); // refreshed but cannot identify the user: stay a guest
      })
      .catch(() => {
        if (!cancelled) writeHint(false); // expired / revoked: back to guest
      })
      .finally(() => {
        if (!cancelled) setIsInitializing(false);
      });
    return () => {
      cancelled = true;
    };
  }, [restoreSession]);

  const login = useCallback(async (req: LoginRequest): Promise<User> => {
    setIsLoading(true);
    try {
      const res = await authService.login(req);
      setAccessToken(res.accessToken);
      setUser(res.user);
      writeHint(true);
      return res.user;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const logout = useCallback(async (): Promise<void> => {
    try {
      await authService.logout();
    } finally {
      setAccessToken(null);
      setUser(null);
      writeHint(false);
    }
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({ user, isAuthenticated: user !== null, isLoading, isInitializing, login, logout }),
    [user, isLoading, isInitializing, login, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
