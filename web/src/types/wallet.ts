// Wallet and wallet-ledger DTOs (data-model.md: wallets, wallet_transactions).

import type { Decimal, ISODateTime } from './common';

export interface Wallet {
  balance: Decimal;
  reserved: Decimal;
  /** Computed: balance − reserved (BR-11). */
  available: Decimal;
}

export type WalletTransactionType =
  | 'deposit'
  | 'withdraw'
  | 'hold'
  | 'release'
  | 'settlement'
  | 'coupon'
  | 'maturity';

export type WalletTransactionStatus = 'done' | 'pending';

export interface WalletTransaction {
  id: number;
  type: WalletTransactionType;
  /** Signed per BR-15 (stored signed; display sign governed by wallet view). */
  amount: Decimal;
  balanceAfter: Decimal;
  refTxnId: number | null;
  status: WalletTransactionStatus;
  createdAt: ISODateTime;
}

export interface DepositRequest {
  amount: Decimal;
}

export interface WithdrawRequest {
  amount: Decimal;
}
