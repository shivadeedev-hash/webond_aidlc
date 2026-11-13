// History service interface + env-selected implementation
// (api-spec.md /api/transactions).

import { DATA_SOURCE } from '../../lib/constants';
import type { HistoryQuery, Transaction, TransactionDetail } from '../../types';
import type { DownloadedFile, Paginated } from '../../types/common';
import { mockHistoryService } from './mock';
import { httpHistoryService } from './http';

export interface HistoryService {
  listTransactions(query?: HistoryQuery): Promise<Paginated<Transaction>>;
  getTransaction(txnRef: string): Promise<TransactionDetail>;
  cancelTransaction(txnRef: string): Promise<Transaction>;
  /** Excel (.xlsx) export of the filtered history (US-028). API mode only. */
  exportTransactions(query?: HistoryQuery): Promise<DownloadedFile>;
  /** Monthly statement PDF for `month` in yyyyMM form (US-028). API mode only. */
  downloadStatement(month: string): Promise<DownloadedFile>;
}

export const historyService: HistoryService =
  DATA_SOURCE === 'api' ? httpHistoryService : mockHistoryService;
