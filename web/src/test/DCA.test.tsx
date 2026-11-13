import { describe, it, expect } from 'vitest';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import { Route, Routes } from 'react-router-dom';
import DCA from '../pages/DCA';
import { renderWithProviders } from './renderApp';

function renderDca(): void {
  renderWithProviders(
    <Routes>
      <Route path="/dca" element={<DCA />} />
    </Routes>,
    { initialEntries: ['/dca'], authenticated: true },
  );
}

// DCA form validation (US-019, BR-06/07).
describe('DCA page', () => {
  it('lists seeded plans in the table', async () => {
    renderDca();
    // Plan assets resolve to catalog codes (K-GLOBAL for fund plan 4001).
    // K-GLOBAL appears both in the table and the round-history selector.
    const matches = await screen.findAllByText('K-GLOBAL');
    expect(matches.length).toBeGreaterThan(0);
  });

  it('rejects an amount below the fund minimum (BR-07)', async () => {
    renderDca();
    // Open the create form.
    fireEvent.click(await screen.findByRole('button', { name: '+ New DCA plan' }));

    const amount = await screen.findByLabelText('Amount per round (THB)');
    fireEvent.change(amount, { target: { value: '100' } }); // below 500 MF minimum

    fireEvent.click(screen.getByRole('button', { name: 'Save plan' }));

    expect(await screen.findByText(/Minimum per round is 500 THB/i)).toBeInTheDocument();
  });

  it('amount field accepts "1,000", drops a minus sign, and quick picks fill it', async () => {
    renderDca();
    fireEvent.click(await screen.findByRole('button', { name: '+ New DCA plan' }));
    const amount = await screen.findByLabelText('Amount per round (THB)');

    fireEvent.change(amount, { target: { value: '฿1,000' } });
    expect(amount).toHaveValue('1,000');
    fireEvent.change(amount, { target: { value: '-250' } });
    expect(amount).toHaveValue('250');

    fireEvent.click(screen.getByRole('button', { name: '5,000' }));
    expect(amount).toHaveValue('5,000');
    // A valid amount saves without the minimum-amount error.
    fireEvent.click(screen.getByRole('button', { name: 'Save plan' }));
    await waitFor(() => {
      expect(screen.queryByText(/Minimum per round/i)).not.toBeInTheDocument();
    });
  });

  it('only offers Mutual Fund and Gold assets (bond excluded, BR-06)', async () => {
    renderDca();
    fireEvent.click(await screen.findByRole('button', { name: '+ New DCA plan' }));

    const assetSelect = (await screen.findByLabelText('Asset')) as HTMLSelectElement;
    await waitFor(() => {
      expect(assetSelect.options.length).toBeGreaterThan(0);
    });
    const optionText = Array.from(assetSelect.options).map((o) => o.textContent ?? '');
    // No bond code should be present in the picker.
    expect(optionText.some((txt) => txt.includes('LB28DA'))).toBe(false);
    // A fund and a gold option are present.
    expect(optionText.some((txt) => txt.includes('K-GLOBAL'))).toBe(true);
    expect(optionText.some((txt) => txt.includes('GOLD-965'))).toBe(true);
  });
});
