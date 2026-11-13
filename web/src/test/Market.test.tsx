import { describe, it, expect } from 'vitest';
import { screen, waitFor, fireEvent } from '@testing-library/react';
import Market from '../pages/Market';
import { renderWithProviders } from './renderApp';

// Market render + filter (US-011). Default locale 'en' via the harness.
describe('Market page', () => {
  it('renders fund cards from the catalog on load', async () => {
    renderWithProviders(<Market />);
    // Seeded funds (funds.json): K-GLOBAL is risk-level 5, appears by default.
    expect(await screen.findByText('K-GLOBAL')).toBeInTheDocument();
    expect(screen.getByText('SCBSET50')).toBeInTheDocument();
  });

  it('switches to the Bond tab and shows bond assets only', async () => {
    renderWithProviders(<Market />);
    await screen.findByText('K-GLOBAL');

    fireEvent.click(screen.getByRole('tab', { name: 'Bond' }));

    // Bond codes appear; a fund code no longer does.
    expect(await screen.findByText('LB28DA')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.queryByText('K-GLOBAL')).not.toBeInTheDocument();
    });
  });

  it('filters the list via the search box', async () => {
    renderWithProviders(<Market />);
    await screen.findByText('K-GLOBAL');

    const search = screen.getByLabelText('Search fund name / code');
    fireEvent.change(search, { target: { value: 'TMBMONEY' } });

    expect(await screen.findByText('TMBMONEY')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.queryByText('K-GLOBAL')).not.toBeInTheDocument();
    });
  });
});
