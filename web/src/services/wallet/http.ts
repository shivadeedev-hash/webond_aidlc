// HTTP wallet implementation — identical interface to the mock (api wave).

import { apiClient } from '../../lib/apiClient';
import type { WalletService } from './index';
import type { DepositRequest, Wallet, WalletTransaction, WithdrawRequest } from '../../types';
import type { Paginated } from '../../types/common';
import { toWalletTransaction, type ApiWalletTransaction } from '../apiMappers';

export const httpWalletService: WalletService = {
  async getWallet(): Promise<Wallet> {
    const { data } = await apiClient.get<Wallet>('/wallet');
    return data;
  },

  async listTransactions(page = 1, pageSize = 20): Promise<Paginated<WalletTransaction>> {
    const { data, meta } = await apiClient.get<ApiWalletTransaction[]>('/wallet/transactions', {
      page,
      pageSize,
    });
    const items = (data ?? []).map(toWalletTransaction);
    return {
      items,
      meta: {
        page: meta?.page ?? page,
        pageSize: meta?.pageSize ?? pageSize,
        total: meta?.total ?? items.length,
        hasMore: meta?.hasMore ?? false,
      },
    };
  },

  async deposit(req: DepositRequest): Promise<Wallet> {
    const { data } = await apiClient.post<Wallet>('/wallet/deposit', req);
    return data;
  },

  async withdraw(req: WithdrawRequest): Promise<Wallet> {
    const { data } = await apiClient.post<Wallet>('/wallet/withdraw', req);
    return data;
  },
};
