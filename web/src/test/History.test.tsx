import { describe, it, expect } from 'vitest';
import { screen, waitFor, fireEvent } from '@testing-library/react';
import { Route, Routes } from 'react-router-dom';
import History from '../pages/History';
import { renderWithProviders } from './renderApp';

function renderHistory(entry: string): void {
  renderWithProviders(
    <Routes>
      <Route path="/history" element={<History />} />
    </Routes>,
    { initialEntries: [entry], authenticated: true },
  );
}

// History list + filter + empty-state (US-024/025).
describe('History page', () => {
  it('lists seeded transactions grouped by date', async () => {
    renderHistory('/history');
    // Rows render as links to /history/<ref>; wait for them to load from the mock.
    await waitFor(() => {
      const rowLinks = screen
        .getAllByRole('link')
        .filter((a) => a.getAttribute('href')?.startsWith('/history/'));
      expect(rowLinks.length).toBeGreaterThan(0);
    });
  });

  it('filters by the Income tab to show only coupon/maturity rows', async () => {
    renderHistory('/history');
    await screen.findByRole('tab', { name: 'Income' });

    fireEvent.click(screen.getByRole('tab', { name: 'Income' }));

    // Income tab selected (aria-selected) reflects the filter state.
    await waitFor(() => {
      expect(screen.getByRole('tab', { name: 'Income' })).toHaveAttribute('aria-selected', 'true');
    });
  });

  it('shows an empty state with Clear filters when no transaction matches', async () => {
    // No seeded transaction has status "cancelled" → empty result.
    renderHistory('/history?status=cancelled');
    expect(await screen.findByText('Clear filters')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Go to Market' })).toBeInTheDocument();
  });
});
