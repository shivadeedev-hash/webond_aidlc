// Converts Go API responses (api/internal/http + service view structs) into the
// web domain types the pages use. The two sides grew separately and differ in
// field names and nesting (e.g. `type` vs `assetType`, flat dashboard totals vs
// `totals`), so every http.ts goes through these pure functions. They are tested
// against real captured API responses (test/fixtures/apiResponses.json).
//
// Rule: never invent data. A value the API does not provide becomes null (or the
// documented neutral default noted next to it).

import type {
  AssetListItem,
  AssetType,
  Dashboard,
  DcaPlan,
  DcaRound,
  DcaStatus,
  Holding,
  InboxMessage,
  Notification,
  NotificationCategory,
  Order,
  OrderMode,
  OrderStatus,
  Portfolio,
  PricePeriod,
  PriceSeries,
  RiskLevel,
  Transaction,
  TransactionDetail,
  TransactionStatus,
  TransactionType,
  UpcomingEvent,
  WalletTransaction,
} from '../types';

// ---------------------------------------------------------------- API shapes

export interface ApiAssetListItem {
  type: AssetType;
  id: number;
  code: string;
  name: string;
  category?: string;
  riskLevel?: number;
  latestPrice?: string;
  bidPrice?: string;
  askPrice?: string;
  ytm?: string;
  maturityDate?: string;
  periodReturn?: string;
  priceAsOf?: string;
}

export interface ApiAssetDetail {
  type: AssetType;
  id: number;
  code: string;
  name: string;
  fields: Record<string, string | number | undefined>;
  priceAsOf?: string;
  priceUnavailable: boolean;
}

export interface ApiPricePoint {
  date: string;
  value: string;
  bid?: string;
  ask?: string;
}

export interface ApiHolding {
  assetType: AssetType;
  assetId: number;
  code: string;
  name: string;
  quantity: string;
  avgCost: string;
  totalCost?: string;
  marketPrice: string;
  marketValue: string;
  unrealizedPl: string;
  unrealizedPct: string;
  nextCouponDate?: string;
  maturityDate?: string;
  accruedInterest?: string;
}

export interface ApiPortfolio {
  summary: { marketValue: string; totalCost: string; unrealizedPl: string; realizedPlYtd: string };
  holdings: ApiHolding[];
}

export interface ApiDashboard {
  totalValue: string;
  totalPl: string;
  cashAvailable: string;
  allocation: { assetType: AssetType; value: string; pct: string }[];
  trend: { date: string; value: string }[];
  topMovers: {
    assetType?: AssetType;
    assetId?: number;
    code: string;
    name?: string;
    pct: string;
  }[];
  dcaSummary: { activePlans: number; totalInvested: string; monthlyCommitment?: string };
  upcomingEvents: { type: string; date: string; ref?: string; amount?: string }[];
  lastUpdated: string;
}

export interface ApiTransaction {
  txnRef: string;
  type: TransactionType;
  assetType?: AssetType;
  assetId?: number;
  amount: string;
  quantity?: string;
  price?: string;
  netAmount: string;
  status: TransactionStatus;
  occurredAt: string;
  realizedPl?: string;
  tradeDate?: string;
  settlementDate?: string;
}

export interface ApiTransactionDetail {
  transaction: ApiTransaction;
  timeline: { status: string; note?: string; changedAt: string }[];
  fee: string;
  accruedInterest?: string;
  dcaPlanId?: number;
  dcaRound?: number;
}

export interface ApiOrder {
  txnRef: string;
  status: string;
  assetType: AssetType;
  assetId: number;
  side: 'buy' | 'sell';
  amount?: string;
  units?: string;
  weightGrams?: string;
  price?: string;
  netAmount?: string;
  realizedPl?: string;
  tradeDate?: string;
  settlementDate?: string;
}

export interface ApiDcaPlan {
  id: number;
  assetType: 'fund' | 'gold';
  assetId: number;
  amountPerRound: string;
  frequency: 'weekly' | 'monthly';
  runDay: number;
  startDate: string;
  nextRunDate?: string | null;
  roundsExecuted: number;
  totalInvested: string;
  status: DcaStatus;
}

export interface ApiDcaRound {
  id: number;
  runDate: string;
  status: 'success' | 'failed';
  failureReason?: string;
}

export interface ApiWalletTransaction {
  id: number;
  type: WalletTransaction['type'];
  amount: string;
  balanceAfter: string;
  status: WalletTransaction['status'];
  createdAt: string;
}

export interface ApiNotification {
  id: number;
  category: NotificationCategory;
  subject: string;
  body: string;
  readAt: string | null;
  createdAt: string;
}

export interface ApiInboxMessage {
  id: number;
  category: NotificationCategory;
  subject: string;
  body: string;
  token?: string;
  createdAt: string;
}

// ------------------------------------------------------------------- helpers

/** Missing / empty string -> null. */
function opt(v: string | undefined | null): string | null {
  return v === undefined || v === null || v === '' ? null : v;
}

/** Risk level clamped to 1-5. Gold products have no risk level in the data
 *  model (gold_products has no risk column), so they get the lowest level (1). */
function risk(v: number | string | undefined): RiskLevel {
  const n = Number(v);
  return (Number.isFinite(n) && n >= 1 && n <= 5 ? Math.round(n) : 1) as RiskLevel;
}

function str(v: string | number | undefined): string {
  return v === undefined ? '' : String(v);
}

function num(v: string | number | undefined, fallback = 0): number {
  const n = Number(v);
  return Number.isFinite(n) ? n : fallback;
}

/** a / b * 100 with 2 decimals; null when b is 0 or a value is missing. */
function pct(a: string, b: string): string | null {
  const x = Number(a);
  const y = Number(b);
  if (!Number.isFinite(x) || !Number.isFinite(y) || y === 0) return null;
  return ((x / y) * 100).toFixed(2);
}

// --------------------------------------------------------------- catalog

export function toAssetListItem(a: ApiAssetListItem): AssetListItem {
  const base = {
    id: a.id,
    code: a.code,
    name: a.name,
    riskLevel: risk(a.riskLevel),
    isActive: true, // the API lists active assets only
    quote: {
      latestPrice: a.latestPrice ?? a.askPrice ?? '',
      bidPrice: a.bidPrice,
      askPrice: a.askPrice,
      pricedAt: a.priceAsOf ?? '',
      periodReturnPct: opt(a.periodReturn),
    },
  };
  // Fields the list does not carry stay empty; the detail call fills them.
  switch (a.type) {
    case 'fund':
      return {
        ...base,
        assetType: 'fund',
        category: a.category ?? '',
        minPurchase: '',
        settlementDays: 0,
        cutoffTime: '',
        factSheetUrl: '',
      };
    case 'bond':
      return {
        ...base,
        assetType: 'bond',
        issuer: '',
        couponRate: '',
        couponFrequency: 0,
        faceValue: '',
        maturityDate: a.maturityDate ?? '',
        ytm: a.ytm ?? '',
        creditRating: '',
        offeringStart: '',
        offeringEnd: '',
        minLot: '',
      };
    case 'gold':
    default:
      return { ...base, assetType: 'gold', purity: '' };
  }
}

export function toAssetDetail(d: ApiAssetDetail): AssetListItem {
  const f = d.fields ?? {};
  const pricedAt = d.priceAsOf ?? '';
  const base = { id: d.id, code: d.code, name: d.name, isActive: true };

  switch (d.type) {
    case 'fund':
      return {
        ...base,
        assetType: 'fund',
        riskLevel: risk(f.riskLevel),
        category: str(f.category),
        minPurchase: str(f.minPurchase),
        settlementDays: num(f.settlementDays),
        cutoffTime: str(f.cutoffTime).slice(0, 5), // "15:30:00" -> "15:30"
        factSheetUrl: str(f.factSheetUrl),
        quote: {
          latestPrice: str(f.nav),
          pricedAt,
          periodReturnPct: null,
          unavailable: d.priceUnavailable,
        },
      };
    case 'bond':
      return {
        ...base,
        assetType: 'bond',
        riskLevel: risk(f.riskLevel),
        issuer: str(f.issuer),
        couponRate: str(f.couponRate),
        couponFrequency: num(f.couponFrequency),
        faceValue: str(f.faceValue),
        maturityDate: str(f.maturityDate),
        ytm: str(f.ytm),
        creditRating: str(f.creditRating),
        offeringStart: str(f.offeringStart),
        offeringEnd: str(f.offeringEnd),
        minLot: str(f.minLot),
        quote: {
          latestPrice: str(f.price), // THB per unit (the API also sends cleanPrice, % of par)
          pricedAt,
          periodReturnPct: null,
          unavailable: d.priceUnavailable,
        },
      };
    case 'gold':
    default:
      return {
        ...base,
        assetType: 'gold',
        riskLevel: risk(undefined),
        purity: str(f.purity),
        quote: {
          latestPrice: str(f.ask),
          bidPrice: opt(str(f.bid)) ?? undefined,
          askPrice: opt(str(f.ask)) ?? undefined,
          pricedAt,
          periodReturnPct: null,
          unavailable: d.priceUnavailable,
        },
      };
  }
}

export function toPriceSeries(
  assetType: AssetType,
  assetId: number,
  period: PricePeriod,
  points: ApiPricePoint[],
): PriceSeries {
  return {
    assetType,
    assetId,
    period,
    points: points.map((p) => ({ date: p.date, value: p.value })),
  };
}

// -------------------------------------------------- portfolio / dashboard

const EVENT_LABEL: Record<string, string> = { coupon: 'Coupon', maturity: 'Maturity' };

export function toHolding(h: ApiHolding, index: number): Holding {
  const isBond = h.assetType === 'bond';
  // For bonds the next event is the next coupon, else maturity.
  const nextEventDate = isBond ? (opt(h.nextCouponDate) ?? opt(h.maturityDate)) : null;
  const nextEventLabel = !nextEventDate
    ? null
    : h.nextCouponDate
      ? EVENT_LABEL.coupon
      : EVENT_LABEL.maturity;
  return {
    id: index + 1, // the API does not expose holding ids; a stable list key is enough
    assetType: h.assetType,
    assetId: h.assetId,
    assetCode: h.code,
    assetName: h.name,
    quantity: h.quantity,
    lockedQuantity: '0', // not returned by the portfolio endpoint
    avgCost: h.avgCost,
    totalCost: h.totalCost ?? (Number(h.avgCost) * Number(h.quantity)).toFixed(2),
    marketPrice: h.marketPrice,
    marketValue: h.marketValue,
    unrealizedPl: h.unrealizedPl,
    unrealizedPlPct: h.unrealizedPct,
    nextEventDate,
    nextEventLabel,
  };
}

export function toPortfolio(p: ApiPortfolio): Portfolio {
  const { summary } = p;
  return {
    summary: {
      totalCost: summary.totalCost,
      totalMarketValue: summary.marketValue,
      totalUnrealizedPl: summary.unrealizedPl,
      totalUnrealizedPlPct: pct(summary.unrealizedPl, summary.totalCost) ?? '0.00',
    },
    holdings: (p.holdings ?? []).map(toHolding),
  };
}

function toEvent(e: ApiDashboard['upcomingEvents'][number]): UpcomingEvent {
  const type: UpcomingEvent['type'] = e.type === 'coupon' || e.type === 'maturity' ? e.type : 'dca';
  const name = type === 'dca' ? 'DCA' : EVENT_LABEL[type];
  return {
    type,
    date: e.date,
    label: e.ref ? `${name} · ${e.ref}` : name,
    amount: opt(e.amount),
  };
}

export function toDashboard(d: ApiDashboard): Dashboard {
  // Unrealized P/L relative to cost (cost = market value - unrealized P/L).
  const cost = (Number(d.totalValue) - Number(d.totalPl)).toFixed(2);
  return {
    totals: {
      totalValue: d.totalValue,
      totalPl: d.totalPl,
      totalPlPct: pct(d.totalPl, cost) ?? '0.00',
    },
    allocation: d.allocation ?? [],
    trend: d.trend ?? [],
    topMovers: (d.topMovers ?? []).map((m) => ({
      assetType: m.assetType ?? 'fund',
      assetId: m.assetId ?? 0,
      assetCode: m.code,
      assetName: m.name ?? m.code,
      changePct: m.pct,
    })),
    dcaSummary: {
      activePlans: d.dcaSummary?.activePlans ?? 0,
      totalInvested: d.dcaSummary?.totalInvested ?? '0.00',
      monthlyCommitment: d.dcaSummary?.monthlyCommitment ?? '0.00',
    },
    upcomingEvents: (d.upcomingEvents ?? []).map(toEvent),
    asOf: d.lastUpdated,
  };
}

// ---------------------------------------------------------------- history

export function toTransaction(t: ApiTransaction, extra: Partial<Transaction> = {}): Transaction {
  return {
    id: 0, // not exposed; the reference number is the identifier
    txnRef: t.txnRef,
    userId: 0, // implicit (always the signed-in user)
    type: t.type,
    assetType: t.assetType ?? null,
    assetId: t.assetId ?? null,
    amount: t.amount,
    quantity: opt(t.quantity),
    price: opt(t.price),
    fee: '0.00', // fees are 0 in phase 1 (BR-17); the detail call returns the real value
    netAmount: t.netAmount,
    status: t.status,
    orderId: null,
    dcaPlanId: null,
    dcaRound: null,
    tradeDate: opt(t.tradeDate),
    settlementDate: opt(t.settlementDate),
    accruedInterest: null,
    realizedPl: opt(t.realizedPl),
    refTxnId: null,
    occurredAt: t.occurredAt,
    createdAt: t.occurredAt,
    ...extra,
  };
}

export function toTransactionDetail(d: ApiTransactionDetail): TransactionDetail {
  return {
    transaction: toTransaction(d.transaction, {
      fee: d.fee,
      accruedInterest: opt(d.accruedInterest),
      dcaPlanId: d.dcaPlanId ?? null,
      dcaRound: d.dcaRound ?? null,
    }),
    statusHistory: (d.timeline ?? []).map((s, i) => ({
      id: i + 1,
      transactionId: 0,
      status: s.status,
      note: s.note ?? '',
      changedBy: '',
      changedAt: s.changedAt,
    })),
  };
}

// ----------------------------------------------------------------- orders

const ORDER_STATUSES: OrderStatus[] = ['pending', 'processing', 'completed', 'failed', 'cancelled'];

export function toOrder(o: ApiOrder): Order {
  const mode: OrderMode = o.weightGrams ? 'weight' : o.side === 'sell' ? 'units' : 'amount';
  return {
    id: 0,
    txnRef: o.txnRef,
    userId: 0,
    assetType: o.assetType,
    assetId: o.assetId,
    side: o.side,
    orderMode: mode,
    amount: opt(o.amount),
    units: opt(o.units),
    weightGrams: opt(o.weightGrams),
    price: opt(o.price),
    fee: '0.00',
    netAmount: opt(o.netAmount),
    status: (ORDER_STATUSES as string[]).includes(o.status) ? (o.status as OrderStatus) : 'pending',
    dcaPlanId: null,
    tradeDate: opt(o.tradeDate),
    settlementDate: opt(o.settlementDate),
    realizedPl: opt(o.realizedPl),
    riskAck: false,
    createdAt: '',
    updatedAt: '',
  };
}

// -------------------------------------------------------------------- DCA

export function toDcaPlan(p: ApiDcaPlan): DcaPlan {
  return {
    id: p.id,
    userId: 0,
    assetType: p.assetType,
    assetId: p.assetId,
    amountPerRound: p.amountPerRound,
    frequency: p.frequency,
    runDay: p.runDay,
    startDate: p.startDate,
    endDate: null,
    maxRounds: null,
    roundsExecuted: p.roundsExecuted,
    totalInvested: p.totalInvested,
    consecutiveFailures: 0,
    nextRunDate: opt(p.nextRunDate ?? undefined),
    status: p.status,
    createdAt: '',
    updatedAt: '',
  };
}

export function toDcaRound(r: ApiDcaRound, planId: number): DcaRound {
  return {
    id: r.id,
    dcaPlanId: planId,
    runDate: r.runDate,
    orderId: null,
    status: r.status,
    failureReason: opt(r.failureReason),
    createdAt: r.runDate,
  };
}

// -------------------------------------------------- wallet / notifications

export function toWalletTransaction(w: ApiWalletTransaction): WalletTransaction {
  return { ...w, refTxnId: null };
}

export function toNotification(n: ApiNotification): Notification {
  return { ...n, channel: 'inapp' };
}

export function toInboxMessage(m: ApiInboxMessage): InboxMessage {
  return {
    id: m.id,
    channel: 'email',
    category: m.category,
    subject: m.subject,
    body: m.body,
    readAt: null,
    createdAt: m.createdAt,
    token: opt(m.token),
    tokenExpiresAt: null,
  };
}
