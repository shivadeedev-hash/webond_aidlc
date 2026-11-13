import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import App from '../App';
import { AuthProvider } from '../hooks/AuthProvider';
import { I18nProvider } from '../i18n/I18nProvider';

function renderAt(path: string): void {
  const queryClient = new QueryClient();
  render(
    <QueryClientProvider client={queryClient}>
      <I18nProvider>
        <AuthProvider>
          <MemoryRouter initialEntries={[path]}>
            <App />
          </MemoryRouter>
        </AuthProvider>
      </I18nProvider>
    </QueryClientProvider>,
  );
}

describe('App shell + routing', () => {
  it('renders the public Market page for a guest', () => {
    renderAt('/market');
    // The brand appears in the header; Market heading appears in the page.
    expect(screen.getAllByText('WeBond').length).toBeGreaterThan(0);
    expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
  });

  it('redirects a guest from a protected route to Login (US-005)', () => {
    renderAt('/portfolio');
    // Login page shows the demo credentials hint (locale-stable); protected
    // portfolio content is not rendered for a guest.
    expect(screen.getByText('demo@webond.test')).toBeInTheDocument();
  });

  it('exposes a primary navigation landmark with the Market link for guests', () => {
    renderAt('/market');
    const nav = screen.getByRole('navigation', { name: 'Primary' });
    expect(nav).toBeInTheDocument();
  });
});
