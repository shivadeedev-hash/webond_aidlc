// DCA plan + round DTOs (data-model.md: dca_plans, dca_rounds). MF & Gold only (BR-06).

import type { Decimal, ISODate, ISODateTime } from './common';

/** DCA is allowed for funds and gold only (bond excluded, BR-06). */
export type DcaAssetType = 'fund' | 'gold';

export type DcaFrequency = 'weekly' | 'monthly';

export type DcaStatus = 'active' | 'paused' | 'completed' | 'cancelled';

export interface DcaPlan {
  id: number;
  userId: number;
  assetType: DcaAssetType;
  assetId: number;
  amountPerRound: Decimal;
  frequency: DcaFrequency;
  /** 1–7 (weekly) or 1–28 (monthly). */
  runDay: number;
  startDate: ISODate;
  endDate: ISODate | null;
  maxRounds: number | null;
  roundsExecuted: number;
  totalInvested: Decimal;
  consecutiveFailures: number;
  nextRunDate: ISODate | null;
  status: DcaStatus;
  createdAt: ISODateTime;
  updatedAt: ISODateTime;
}

export type DcaRoundStatus = 'success' | 'failed';

export interface DcaRound {
  id: number;
  dcaPlanId: number;
  runDate: ISODate;
  orderId: number | null;
  status: DcaRoundStatus;
  failureReason: string | null;
  createdAt: ISODateTime;
}

export interface CreateDcaPlanRequest {
  assetType: DcaAssetType;
  assetId: number;
  amountPerRound: Decimal;
  frequency: DcaFrequency;
  runDay: number;
  startDate: ISODate;
  endDate?: ISODate;
  maxRounds?: number;
}

export interface UpdateDcaPlanRequest {
  amountPerRound?: Decimal;
  runDay?: number;
}
