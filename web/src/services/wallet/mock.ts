// Mock wallet implementation backed by mocks/wallet.json. Deposit/withdraw update
// an in-memory copy so the demo reflects actions within a session.

import walletData from '../../mocks/wallet.json';
import { ApiClientError } from '../../lib/errors';
import { mockResolve, paginate } from '../mockUtils';
import type { WalletService } from './index';
import type { DepositRequest, Wallet, WalletTransaction, WithdrawRequest } from '../../types';

const wallet: Wallet = structuredClone(walletData.wallet) as Wallet;
const transactions: WalletTransaction[] = structuredClone(
  walletData.transactions,
) as WalletTransaction[];

function recompute(): void {
  wallet.available = (Number(wallet.balance) - Number(wallet.reserved)).toFixed(4);
}

export const mockWalletService: WalletService = {
  async getWallet(): Promise<Wallet> {
    return mockResolve(wallet);
  },

  async listTransactions(page = 1, pageSize = 20) {
    const sorted = [...transactions].sort(
      (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
    );
    return mockResolve(paginate(sorted, page, pageSize));
  },

  async deposit(req: DepositRequest): Promise<Wallet> {
    const amount = Number(req.amount);
    if (!(amount > 0)) throw new ApiClientError('VALIDATION_ERROR', 'Enter a valid amount.');
    wallet.balance = (Number(wallet.balance) + amount).toFixed(4);
    recompute();
    return mockResolve(wallet);
  },

  async withdraw(req: WithdrawRequest): Promise<Wallet> {
    const amount = Number(req.amount);
    if (!(amount > 0)) throw new ApiClientError('VALIDATION_ERROR', 'Enter a valid amount.');
    if (amount > Number(wallet.available)) {
      throw new ApiClientError('INSUFFICIENT_BALANCE');
    }
    wallet.balance = (Number(wallet.balance) - amount).toFixed(4);
    recompute();
    return mockResolve(wallet);
  },
};
