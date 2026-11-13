// Contract tests: real responses captured from the running Go API
// (fixtures/apiResponses.json) must map into the shapes the pages read.
// Before these mappers existed, http.ts passed the raw JSON through and the
// Portfolio and Dashboard pages crashed in API mode.
import { describe, expect, it } from 'vitest';
import api from './fixtures/apiResponses.json';
import {
  toAssetDetail,
  toAssetListItem,
  toDashboard,
  toDcaPlan,
  toDcaRound,
  toNotification,
  toOrder,
  toPortfolio,
  toPriceSeries,
  toTransaction,
  toTransactionDetail,
  toWalletTransaction,
  type ApiAssetDetail,
  type ApiAssetListItem,
  type ApiDashboard,
  type ApiDcaPlan,
  type ApiDcaRound,
  type ApiNotification,
  type ApiPortfolio,
  type ApiPricePoint,
  type ApiTransaction,
  type ApiTransactionDetail,
  type ApiWalletTransaction,
} from '../services/apiMappers';

const data = <T>(key: keyof typeof api): T => (api[key] as unknown as { data: T }).data;

describe('catalog', () => {
  it('maps list items to assetType + quote', () => {
    const funds = data<ApiAssetListItem[]>('assetsFund').map(toAssetListItem);
    const fund = funds.find((f) => f.code === 'K-GLOBAL');
    expect(fund?.assetType).toBe('fund');
    expect(fund?.riskLevel).toBe(5);
    expect(fund?.quote.latestPrice).toMatch(/^\d+\.\d{4}$/);
    expect(fund?.quote.pricedAt).not.toBe('');

    const bond = data<ApiAssetListItem[]>('assetsBond').map(toAssetListItem)[0];
    expect(bond.assetType).toBe('bond');
    // THB per unit (face 1,000), never the raw % of par.
    expect(Number(bond.quote.latestPrice)).toBeGreaterThan(500);

    const gold = data<ApiAssetListItem[]>('assetsGold').map(toAssetListItem)[0];
    expect(gold.assetType).toBe('gold');
    expect(gold.quote.bidPrice).toBeTruthy();
    expect(gold.quote.askPrice).toBeTruthy();
    expect(gold.riskLevel).toBe(1); // gold has no risk level in the data model
  });

  it('flattens detail fields per asset type', () => {
    const fund = toAssetDetail(data<ApiAssetDetail>('detailFund'));
    expect(fund).toMatchObject({
      assetType: 'fund',
      code: 'K-GLOBAL',
      minPurchase: '500.0000',
      cutoffTime: '15:30',
    });
    if (fund.assetType === 'fund') expect(fund.settlementDays).toBeGreaterThan(0);

    const bond = toAssetDetail(data<ApiAssetDetail>('detailBond'));
    expect(bond).toMatchObject({
      assetType: 'bond',
      faceValue: '1000.0000',
      minLot: '100000.0000',
    });
    expect(Number(bond.quote.latestPrice)).toBeGreaterThan(500);

    const gold = toAssetDetail(data<ApiAssetDetail>('detailGold'));
    expect(gold.assetType).toBe('gold');
    expect(gold.quote.askPrice).toBe(gold.quote.latestPrice);
    expect(typeof gold.quote.unavailable).toBe('boolean');
  });

  it('wraps price points into a series', () => {
    const pts = data<ApiPricePoint[]>('pricesBond');
    const s = toPriceSeries('bond', 1, '1M', pts);
    expect(s).toMatchObject({ assetType: 'bond', assetId: 1, period: '1M' });
    expect(s.points).toHaveLength(pts.length);
    expect(Number(s.points[0].value)).toBeGreaterThan(500);
  });
});

describe('portfolio and dashboard (the pages that crashed)', () => {
  it('portfolio summary and holdings have every field the page reads', () => {
    const p = toPortfolio(data<ApiPortfolio>('portfolio'));
    expect(p.summary.totalMarketValue).toMatch(/^\d+\.\d{2}$/);
    expect(p.summary.totalCost).toMatch(/^\d+\.\d{2}$/);
    expect(p.summary.totalUnrealizedPl).toBeDefined();
    expect(p.summary.totalUnrealizedPlPct).toMatch(/^-?\d+\.\d{2}$/);
    expect(p.holdings.length).toBeGreaterThan(0);
    for (const h of p.holdings) {
      expect(h.assetCode).not.toBe('');
      expect(h.assetName).not.toBe('');
      expect(h.unrealizedPlPct).toBeDefined();
      expect(h.totalCost).toBeDefined();
    }
    const bond = p.holdings.find((h) => h.assetType === 'bond');
    expect(bond?.marketValue).toBe('100500.00');
    expect(bond?.nextEventDate).not.toBeNull();
  });

  it('dashboard totals, movers, DCA summary and events', () => {
    const d = toDashboard(data<ApiDashboard>('dashboard'));
    expect(d.totals.totalValue).toMatch(/^\d+\.\d{2}$/);
    expect(d.totals.totalPlPct).toMatch(/^-?\d+\.\d{2}$/);
    expect(d.allocation.length).toBeGreaterThan(0);
    expect(d.trend.length).toBeGreaterThan(0);
    expect(d.topMovers[0]).toMatchObject({
      assetCode: expect.any(String),
      assetName: expect.any(String),
    });
    expect(d.topMovers[0].assetId).toBeGreaterThan(0);
    expect(d.dcaSummary.monthlyCommitment).toMatch(/^\d+\.\d{2}$/);
    for (const e of d.upcomingEvents) {
      expect(['coupon', 'maturity', 'dca']).toContain(e.type);
      expect(e.label).not.toBe('');
    }
    expect(d.asOf).not.toBe('');
  });
});

describe('history, orders, DCA, wallet, notifications', () => {
  it('transactions and detail timeline', () => {
    const list = data<ApiTransaction[]>('transactions').map((t) => toTransaction(t));
    expect(list[0].txnRef).not.toBe('');
    expect(list[0].occurredAt).not.toBe('');
    const detail = toTransactionDetail(data<ApiTransactionDetail>('transactionDetail'));
    expect(detail.statusHistory.length).toBeGreaterThan(0);
    expect(detail.statusHistory[0]).toMatchObject({
      id: 1,
      status: expect.any(String),
      changedAt: expect.any(String),
    });
    expect(detail.transaction.fee).toBeDefined();
  });

  it('order results', () => {
    const o = toOrder({
      txnRef: 'WB2611120000001',
      status: 'completed',
      assetType: 'fund',
      assetId: 1,
      side: 'buy',
      amount: '5000.00',
      units: '335.3454',
      price: '14.9100',
      netAmount: '5000.00',
    });
    expect(o).toMatchObject({
      txnRef: 'WB2611120000001',
      status: 'completed',
      orderMode: 'amount',
      weightGrams: null,
    });
    expect(
      toOrder({ ...o, side: 'buy', weightGrams: '1.0000', status: 'weird' } as never).orderMode,
    ).toBe('weight');
  });

  it('DCA plans and rounds', () => {
    const plan = toDcaPlan(data<ApiDcaPlan[]>('dcaPlans')[0]);
    expect(plan).toMatchObject({
      id: expect.any(Number),
      status: expect.any(String),
      amountPerRound: expect.any(String),
    });
    const round = toDcaRound(data<ApiDcaRound[]>('dcaRounds')[0], plan.id);
    expect(round.dcaPlanId).toBe(plan.id);
    expect(round.failureReason).toBeNull(); // "" from the API becomes null
  });

  it('wallet ledger and notifications', () => {
    const w = toWalletTransaction(data<ApiWalletTransaction[]>('walletTx')[0]);
    expect(w.refTxnId).toBeNull();
    expect(w.balanceAfter).toBeDefined();
    const n = toNotification(data<ApiNotification[]>('notifications')[0]);
    expect(n.channel).toBe('inapp');
  });
});
