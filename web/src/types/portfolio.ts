// Holdings, portfolio and dashboard DTOs
// (data-model.md: holdings; FR-21 portfolio, FR-22 dashboard).

import type { AssetType, Decimal, ISODate, ISODateTime } from './common';

export interface Holding {
  id: number;
  assetType: AssetType;
  assetId: number;
  /** Denormalized for display. */
  assetCode: string;
  assetName: string;
  /** Units (fund), bond units, or grams (gold). */
  quantity: Decimal;
  lockedQuantity: Decimal;
  /** Average cost per unit/gram (BR-10). */
  avgCost: Decimal;
  totalCost: Decimal;
  /** Latest market price per unit/gram. */
  marketPrice: Decimal;
  marketValue: Decimal;
  /** Unrealized P/L amount and percentage. */
  unrealizedPl: Decimal;
  unrealizedPlPct: Decimal;
  /** Next coupon/maturity date for bonds, else null (FR-21). */
  nextEventDate: ISODate | null;
  nextEventLabel: string | null;
}

export interface PortfolioSummary {
  totalCost: Decimal;
  totalMarketValue: Decimal;
  totalUnrealizedPl: Decimal;
  totalUnrealizedPlPct: Decimal;
}

export interface Portfolio {
  summary: PortfolioSummary;
  holdings: Holding[];
}

export type TrendPeriod = '1M' | '6M' | '1Y' | '3Y' | '5Y' | 'all';

export interface AllocationSlice {
  assetType: AssetType;
  value: Decimal;
  pct: Decimal;
}

export interface TrendPoint {
  date: ISODate;
  value: Decimal;
}

export interface TopMover {
  assetType: AssetType;
  assetId: number;
  assetCode: string;
  assetName: string;
  changePct: Decimal;
}

export interface DcaSummary {
  activePlans: number;
  totalInvested: Decimal;
  monthlyCommitment: Decimal;
}

export interface UpcomingEvent {
  type: 'coupon' | 'maturity' | 'dca';
  date: ISODate;
  label: string;
  amount: Decimal | null;
}

export interface DashboardTotals {
  totalValue: Decimal;
  totalPl: Decimal;
  totalPlPct: Decimal;
}

export interface Dashboard {
  totals: DashboardTotals;
  allocation: AllocationSlice[];
  trend: TrendPoint[];
  topMovers: TopMover[];
  dcaSummary: DcaSummary;
  upcomingEvents: UpcomingEvent[];
  asOf: ISODateTime;
}
