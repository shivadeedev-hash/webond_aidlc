// DCA service interface + env-selected implementation (api-spec.md /api/dca-plans).

import { DATA_SOURCE } from '../../lib/constants';
import type { CreateDcaPlanRequest, DcaPlan, DcaRound, UpdateDcaPlanRequest } from '../../types';
import { mockDcaService } from './mock';
import { httpDcaService } from './http';

export interface DcaService {
  listPlans(): Promise<DcaPlan[]>;
  createPlan(req: CreateDcaPlanRequest): Promise<DcaPlan>;
  updatePlan(id: number, req: UpdateDcaPlanRequest): Promise<DcaPlan>;
  pausePlan(id: number): Promise<DcaPlan>;
  resumePlan(id: number): Promise<DcaPlan>;
  cancelPlan(id: number): Promise<void>;
  listRounds(id: number): Promise<DcaRound[]>;
}

export const dcaService: DcaService = DATA_SOURCE === 'api' ? httpDcaService : mockDcaService;
