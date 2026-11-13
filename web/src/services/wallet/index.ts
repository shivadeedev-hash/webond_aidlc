// Wallet service interface + env-selected implementation (api-spec.md /api/wallet).

import { DATA_SOURCE } from '../../lib/constants';
import type { DepositRequest, Wallet, WalletTransaction, WithdrawRequest } from '../../types';
import type { Paginated } from '../../types/common';
import { mockWalletService } from './mock';
import { httpWalletService } from './http';

export interface WalletService {
  getWallet(): Promise<Wallet>;
  listTransactions(page?: number, pageSize?: number): Promise<Paginated<WalletTransaction>>;
  deposit(req: DepositRequest): Promise<Wallet>;
  withdraw(req: WithdrawRequest): Promise<Wallet>;
}

export const walletService: WalletService =
  DATA_SOURCE === 'api' ? httpWalletService : mockWalletService;
