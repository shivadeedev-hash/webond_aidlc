// Unified transaction history + status timeline DTOs
// (data-model.md: transactions, transaction_status_history).

import type { AssetType, Decimal, ISODate, ISODateTime } from './common';

export type TransactionType =
  | 'buy'
  | 'sell'
  | 'dca_buy'
  | 'coupon'
  | 'maturity'
  | 'deposit'
  | 'withdraw';

export type TransactionStatus = 'submitted' | 'processing' | 'completed' | 'failed' | 'cancelled';

export interface Transaction {
  id: number;
  txnRef: string;
  userId: number;
  type: TransactionType;
  assetType: AssetType | null;
  assetId: number | null;
  /** Signed per BR-15. */
  amount: Decimal;
  quantity: Decimal | null;
  price: Decimal | null;
  fee: Decimal;
  netAmount: Decimal;
  status: TransactionStatus;
  orderId: number | null;
  dcaPlanId: number | null;
  dcaRound: number | null;
  tradeDate: ISODate | null;
  settlementDate: ISODate | null;
  accruedInterest: Decimal | null;
  realizedPl: Decimal | null;
  refTxnId: number | null;
  occurredAt: ISODateTime;
  createdAt: ISODateTime;
}

export interface TransactionStatusHistoryItem {
  id: number;
  transactionId: number;
  status: string;
  note: string;
  changedBy: string;
  changedAt: ISODateTime;
}

/** Transaction detail = the row + its status timeline (FR-18). */
export interface TransactionDetail {
  transaction: Transaction;
  statusHistory: TransactionStatusHistoryItem[];
}

/** History tabs (api-spec.md GET /api/transactions). */
export type HistoryTab = 'all' | 'orders' | 'dca' | 'cash' | 'income';

export type HistorySort = 'date' | 'amount';
export type SortOrder = 'asc' | 'desc';

/** Query params for the history list (kept in URL, FR-17). */
export interface HistoryQuery {
  tab?: HistoryTab;
  assetType?: AssetType;
  type?: TransactionType;
  status?: TransactionStatus;
  from?: ISODate;
  to?: ISODate;
  q?: string;
  sort?: HistorySort;
  order?: SortOrder;
  page?: number;
  pageSize?: number;
}
