// Shared test harness: wraps a UI tree with the same providers used in main.tsx
// (QueryClient, I18n, Auth, Router). `locale="en"` keeps assertions readable.
// Helper components are declared inside the factory so this module's only export
// is the (non-component) helper — keeps react-refresh/only-export-components happy.

import { useEffect, type ReactElement, type ReactNode } from 'react';
import { render, type RenderResult } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider } from '../hooks/AuthProvider';
import { useAuth } from '../hooks/useAuth';
import { I18nProvider } from '../i18n/I18nProvider';
import demoData from '../mocks/user.json';

export function renderWithProviders(
  ui: ReactElement,
  {
    route = '/',
    initialEntries,
    authenticated = false,
  }: { route?: string; initialEntries?: string[]; authenticated?: boolean } = {},
): RenderResult {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  /** Logs in with the demo credentials on mount (mock auth accepts them). */
  function AutoLogin({ children }: { children: ReactNode }): JSX.Element {
    const { login, isAuthenticated } = useAuth();
    useEffect(() => {
      if (!isAuthenticated) {
        void login({
          email: demoData.demoCredentials.email,
          password: demoData.demoCredentials.password,
        });
      }
    }, [login, isAuthenticated]);
    return <>{children}</>;
  }

  function Wrapper({ children }: { children: ReactNode }): JSX.Element {
    const inner = authenticated ? <AutoLogin>{children}</AutoLogin> : children;
    return (
      <QueryClientProvider client={queryClient}>
        <I18nProvider initialLocale="en">
          <AuthProvider>
            <MemoryRouter initialEntries={initialEntries ?? [route]}>{inner}</MemoryRouter>
          </AuthProvider>
        </I18nProvider>
      </QueryClientProvider>
    );
  }

  return render(ui, { wrapper: Wrapper });
}
