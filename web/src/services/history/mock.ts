// Mock history implementation backed by transactions.json. Supports the history
// tabs/filters/search/sort (FR-16..20) and cancel-pending (BR-13/EX-13).

import txnData from '../../mocks/transactions.json';
import { ApiClientError } from '../../lib/errors';
import { mockResolve, paginate } from '../mockUtils';
import type { DownloadedFile } from '../../types/common';
import type { HistoryService } from './index';
import type {
  HistoryQuery,
  HistoryTab,
  Transaction,
  TransactionDetail,
  TransactionStatusHistoryItem,
  TransactionType,
} from '../../types';

const transactions = structuredClone(txnData.transactions) as Transaction[];
const statusHistory = txnData.statusHistory as Record<string, TransactionStatusHistoryItem[]>;

const TAB_TYPES: Record<Exclude<HistoryTab, 'all'>, TransactionType[]> = {
  orders: ['buy', 'sell'],
  dca: ['dca_buy'],
  cash: ['deposit', 'withdraw'],
  income: ['coupon', 'maturity'],
};

export const mockHistoryService: HistoryService = {
  async listTransactions(query: HistoryQuery = {}) {
    let rows = [...transactions];

    if (query.tab && query.tab !== 'all') {
      const types = TAB_TYPES[query.tab];
      rows = rows.filter((t) => types.includes(t.type));
    }
    if (query.type) rows = rows.filter((t) => t.type === query.type);
    if (query.assetType) rows = rows.filter((t) => t.assetType === query.assetType);
    if (query.status) rows = rows.filter((t) => t.status === query.status);
    if (query.from) rows = rows.filter((t) => t.occurredAt >= query.from!);
    if (query.to) rows = rows.filter((t) => t.occurredAt <= `${query.to!}T23:59:59Z`);
    if (query.q) {
      const q = query.q.toLowerCase();
      rows = rows.filter((t) => t.txnRef.toLowerCase().includes(q));
    }

    const dir = query.order === 'asc' ? 1 : -1;
    rows.sort((a, b) => {
      if (query.sort === 'amount') return (Number(a.amount) - Number(b.amount)) * dir;
      return (new Date(a.occurredAt).getTime() - new Date(b.occurredAt).getTime()) * dir;
    });

    return mockResolve(paginate(rows, query.page, query.pageSize));
  },

  async getTransaction(txnRef: string): Promise<TransactionDetail> {
    const transaction = transactions.find((t) => t.txnRef === txnRef);
    if (!transaction) throw new ApiClientError('TXN_NOT_FOUND', undefined, { status: 404 });
    return mockResolve<TransactionDetail>({
      transaction,
      statusHistory: statusHistory[txnRef] ?? [],
    });
  },

  async cancelTransaction(txnRef: string): Promise<Transaction> {
    const transaction = transactions.find((t) => t.txnRef === txnRef);
    if (!transaction) throw new ApiClientError('TXN_NOT_FOUND', undefined, { status: 404 });
    // Cancellable only while the order is still open (BR-13/EX-13).
    if (transaction.status !== 'submitted' && transaction.status !== 'processing') {
      throw new ApiClientError('CANCEL_NOT_ALLOWED');
    }
    transaction.status = 'cancelled';
    return mockResolve(transaction);
  },

  // Files are produced by the API; the UI keeps these actions disabled in mock mode.
  async exportTransactions(): Promise<DownloadedFile> {
    throw new ApiClientError('UNKNOWN', 'Export is available once the live API is connected.');
  },

  async downloadStatement(): Promise<DownloadedFile> {
    throw new ApiClientError('UNKNOWN', 'Statements are available once the live API is connected.');
  },
};
