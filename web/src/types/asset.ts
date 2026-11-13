// Asset master + price/NAV history DTOs
// (data-model.md: funds/fund_navs, bonds/bond_prices/bond_coupon_schedules,
//  gold_products/gold_prices).

import type { AssetType, Decimal, ISODate, ISODateTime } from './common';
import type { RiskLevel } from './user';

export type { AssetType };

/** Fields common to every tradable asset in catalog listings. */
export interface AssetBase {
  id: number;
  assetType: AssetType;
  code: string;
  name: string;
  riskLevel: RiskLevel;
  isActive: boolean;
}

export interface Fund extends AssetBase {
  assetType: 'fund';
  category: string;
  minPurchase: Decimal;
  /** T+n settlement days. */
  settlementDays: number;
  /** Order cut-off time, HH:mm (BR-02). */
  cutoffTime: string;
  factSheetUrl: string;
}

export interface Bond extends AssetBase {
  assetType: 'bond';
  issuer: string;
  couponRate: Decimal;
  /** Coupons per year. */
  couponFrequency: number;
  faceValue: Decimal;
  maturityDate: ISODate;
  ytm: Decimal;
  creditRating: string;
  offeringStart: ISODate;
  offeringEnd: ISODate;
  /** Add-on: default 100,000 (multiples of 100,000). */
  minLot: Decimal;
}

export interface GoldProduct extends AssetBase {
  assetType: 'gold';
  purity: string;
}

export type Asset = Fund | Bond | GoldProduct;

/** Alias kept for readability; gold master record (data-model.md gold_products). */
export type Gold = GoldProduct;

/** Latest price/NAV + computed period return attached to a catalog row. */
export interface AssetQuote {
  /** NAV for funds, clean price for bonds, ask/mid for gold. */
  latestPrice: Decimal;
  /** Gold BID (sell side); only present for gold. */
  bidPrice?: Decimal;
  /** Gold ASK (buy side); only present for gold. */
  askPrice?: Decimal;
  pricedAt: ISODateTime;
  /** Percentage return over the requested period, e.g. "3.25". */
  periodReturnPct: Decimal | null;
  /** True when the API reports the price as stale/missing (EX-10): trading is blocked. */
  unavailable?: boolean;
}

/** A catalog list item: asset master + its latest quote. */
export type AssetListItem = Asset & { quote: AssetQuote };

/** A single point in a price/NAV history series. */
export interface PricePoint {
  date: ISODate;
  value: Decimal;
}

export interface PriceSeries {
  assetType: AssetType;
  assetId: number;
  period: PricePeriod;
  points: PricePoint[];
}

export type PricePeriod = '1M' | '6M' | '1Y' | '3Y' | '5Y' | 'all';

/** Return-filter periods for the market list (US-011 Add-on). */
export type ReturnPeriod = '3M' | '6M' | '1Y' | '3Y' | '5Y';

/** Bond coupon schedule entry (upcoming/next-event display). */
export interface CouponScheduleItem {
  couponDate: ISODate;
  amountPerUnit: Decimal;
  paid: boolean;
}
