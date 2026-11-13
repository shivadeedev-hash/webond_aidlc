import { describe, it, expect } from 'vitest';
import { screen, fireEvent, within } from '@testing-library/react';
import { Route, Routes } from 'react-router-dom';
import AssetDetail from '../pages/AssetDetail';
import { renderWithProviders } from './renderApp';

function renderDetail(path: string, authenticated: boolean): void {
  renderWithProviders(
    <Routes>
      <Route path="/assets/:type/:id" element={<AssetDetail />} />
    </Routes>,
    { initialEntries: [path], authenticated },
  );
}

const BUY_LABEL = 'How much do you want to buy?';
const SELL_LABEL = 'How much do you want to sell?';

// Asset Detail order flow (US-012..018, Add-on 2026-11-12d).
describe('AssetDetail order flow', () => {
  it('shows the asset detail and, for a guest, a sign-in prompt instead of the panel', async () => {
    renderDetail('/assets/fund/1', false);
    expect(await screen.findByRole('heading', { name: /K-GLOBAL/ })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Log in' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Review order' })).not.toBeInTheDocument();
  });

  it('accepts "1,000" with a comma and places a fund buy order', async () => {
    // Fund 3 (TMBMONEY) is risk-level 1 ≤ the demo user's risk 4 → no risk ack.
    renderDetail('/assets/fund/3', true);
    await screen.findByRole('heading', { name: /TMBMONEY/ });
    const amount = await screen.findByLabelText(BUY_LABEL);
    await screen.findByText(/147,300/);

    fireEvent.change(amount, { target: { value: '฿1,000' } });
    expect(amount).toHaveValue('1,000');
    fireEvent.click(screen.getByRole('button', { name: 'Review order' }));
    expect(await screen.findByText(/Order placed successfully/)).toBeInTheDocument();
  });

  it('never accepts a minus sign and groups thousands while typing', async () => {
    renderDetail('/assets/fund/3', true);
    const amount = await screen.findByLabelText(BUY_LABEL);
    fireEvent.change(amount, { target: { value: '-25000.555' } });
    expect(amount).toHaveValue('25,000.55');
  });

  it('quick picks fill the amount', async () => {
    renderDetail('/assets/fund/3', true);
    const amount = await screen.findByLabelText(BUY_LABEL);
    await screen.findByText(/147,300/);
    fireEvent.click(screen.getByRole('button', { name: '฿5,000.00' }));
    expect(amount).toHaveValue('5,000');
  });

  it('says exactly why a buy is blocked: more than available cash', async () => {
    renderDetail('/assets/fund/3', true);
    const amount = await screen.findByLabelText(BUY_LABEL);
    await screen.findByText(/147,300/);
    fireEvent.change(amount, { target: { value: '99999999' } });
    expect(await screen.findByText(/More than your available cash \(฿147,300/)).toBeInTheDocument();
  });

  it('requires a risk acknowledgement when the asset risk exceeds the user risk (EX-12)', async () => {
    // Fund 1 (K-GLOBAL) is risk-level 5 > the demo user's risk 4.
    renderDetail('/assets/fund/1', true);
    const amount = await screen.findByLabelText(BUY_LABEL);
    await screen.findByText(/147,300/);
    expect(screen.getByText(/Fund risk 5 > your risk 4/)).toBeInTheDocument();

    fireEvent.change(amount, { target: { value: '1000' } });
    fireEvent.click(screen.getByRole('button', { name: 'Review order' }));
    expect(await screen.findByText(/acknowledge the risk notice/i)).toBeInTheDocument();

    fireEvent.click(screen.getByLabelText('I acknowledge the risk'));
    fireEvent.click(screen.getByRole('button', { name: 'Review order' }));
    expect(await screen.findByText(/Order placed successfully/)).toBeInTheDocument();
  });

  it('sell shows the holding, can sell by units, and blocks more than held', async () => {
    // Demo holds 1,256.4310 units of fund 2 (SCBSET50).
    renderDetail('/assets/fund/2', true);
    await screen.findByLabelText(BUY_LABEL);
    fireEvent.click(screen.getByRole('button', { name: 'Sell' }));
    expect(await screen.findByText(/1,256\.4310 units/)).toBeInTheDocument();

    const sellBy = screen.getByRole('group', { name: 'Sell by' });
    fireEvent.click(within(sellBy).getByRole('button', { name: 'In units' }));
    const qty = screen.getByLabelText(SELL_LABEL);
    fireEvent.change(qty, { target: { value: '5000' } });
    expect(await screen.findByText(/You can sell up to 1,256\.4310 units/)).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'All' }));
    expect(qty).toHaveValue('1,256.431');
    fireEvent.click(screen.getByRole('button', { name: 'Review order' }));
    expect(await screen.findByText(/Order placed successfully/)).toBeInTheDocument();
  });

  it('sell is disabled with a clear note when the user holds none', async () => {
    renderDetail('/assets/fund/3', true);
    await screen.findByLabelText(BUY_LABEL);
    fireEvent.click(screen.getByRole('button', { name: 'Sell' }));
    expect(await screen.findByLabelText(SELL_LABEL)).toBeDisabled();
  });
});
