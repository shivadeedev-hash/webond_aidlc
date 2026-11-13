// HTTP history implementation — identical interface to the mock (api wave).

import { apiClient } from '../../lib/apiClient';
import type { HistoryService } from './index';
import type { HistoryQuery, Transaction, TransactionDetail } from '../../types';
import type { DownloadedFile, Paginated } from '../../types/common';
import {
  toTransaction,
  toTransactionDetail,
  type ApiTransaction,
  type ApiTransactionDetail,
} from '../apiMappers';

export const httpHistoryService: HistoryService = {
  async listTransactions(query: HistoryQuery = {}): Promise<Paginated<Transaction>> {
    const { data, meta } = await apiClient.get<ApiTransaction[]>('/transactions', { ...query });
    const items = (data ?? []).map((t) => toTransaction(t));
    return {
      items,
      meta: {
        page: meta?.page ?? query.page ?? 1,
        pageSize: meta?.pageSize ?? query.pageSize ?? 20,
        total: meta?.total ?? items.length,
        hasMore: meta?.hasMore ?? false,
      },
    };
  },

  async getTransaction(txnRef: string): Promise<TransactionDetail> {
    const { data } = await apiClient.get<ApiTransactionDetail>(`/transactions/${txnRef}`);
    return toTransactionDetail(data);
  },

  async cancelTransaction(txnRef: string): Promise<Transaction> {
    // The API answers { status: "cancelled" }; re-read the transaction so the
    // caller gets the full, updated record.
    await apiClient.post(`/transactions/${txnRef}/cancel`);
    return (await this.getTransaction(txnRef)).transaction;
  },
  async exportTransactions(query: HistoryQuery = {}): Promise<DownloadedFile> {
    // Same filters as the list, but never paged: the export covers the whole set.
    const filters: HistoryQuery = { ...query };
    delete filters.page;
    delete filters.pageSize;
    const today = new Date().toISOString().slice(0, 10).replace(/-/g, '');
    return apiClient.download('/transactions/export', {
      query: { ...filters },
      fallbackFilename: `webond-transactions-${today}.xlsx`,
    });
  },
  async downloadStatement(month: string): Promise<DownloadedFile> {
    return apiClient.download(`/statements/${month}`, {
      fallbackFilename: `webond-statement-${month}.pdf`,
    });
  },
};
