import { describe, it, expect } from 'vitest';
import { catalogService } from '../services/catalog';
import { walletService } from '../services/wallet';
import { portfolioService } from '../services/portfolio';
import { historyService } from '../services/history';
import { authService } from '../services/auth';
import { ApiClientError } from '../lib/errors';

// VITE_DATA_SOURCE is unset in tests → defaults to 'mock'. These smoke tests
// confirm the env-selected mock implementations read the seed JSON correctly.

describe('mock service wiring (default VITE_DATA_SOURCE=mock)', () => {
  it('catalog lists seeded assets with quotes', async () => {
    const page = await catalogService.listAssets();
    expect(page.items.length).toBeGreaterThan(0);
    expect(page.items[0].quote).toBeDefined();
  });

  it('catalog filters by type', async () => {
    const funds = await catalogService.listAssets({ type: 'fund' });
    expect(funds.items.every((a) => a.assetType === 'fund')).toBe(true);
  });

  it('wallet exposes balance/reserved/available', async () => {
    const wallet = await walletService.getWallet();
    expect(wallet.balance).toBeTruthy();
    expect(wallet.reserved).toBeTruthy();
    expect(wallet.available).toBeTruthy();
  });

  it('wallet withdraw beyond available throws INSUFFICIENT_BALANCE', async () => {
    await expect(walletService.withdraw({ amount: '99999999.0000' })).rejects.toBeInstanceOf(
      ApiClientError,
    );
  });

  it('portfolio returns a summary + holdings across asset types', async () => {
    const portfolio = await portfolioService.getPortfolio();
    expect(portfolio.holdings.length).toBeGreaterThan(0);
    const types = new Set(portfolio.holdings.map((h) => h.assetType));
    expect(types.size).toBeGreaterThanOrEqual(2);
  });

  it('history lists transactions and supports tab filtering', async () => {
    const all = await historyService.listTransactions();
    expect(all.items.length).toBeGreaterThan(0);
    const income = await historyService.listTransactions({ tab: 'income' });
    expect(income.items.every((t) => t.type === 'coupon' || t.type === 'maturity')).toBe(true);
  });

  it('auth login accepts the demo credentials and rejects others', async () => {
    const res = await authService.login({ email: 'demo@webond.test', password: 'Demo@1234' });
    expect(res.accessToken).toBeTruthy();
    expect(res.user.email).toBe('demo@webond.test');
    await expect(authService.login({ email: 'x@y.z', password: 'bad' })).rejects.toBeInstanceOf(
      ApiClientError,
    );
  });
});
