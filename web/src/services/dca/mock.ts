// Mock DCA implementation backed by mocks/dca.json. Enforces the phase-1 business
// rules relevant to the UI: bond excluded (BR-06), min 500 MF / 100 gold (BR-07).

import dcaData from '../../mocks/dca.json';
import { ApiClientError } from '../../lib/errors';
import { mockResolve } from '../mockUtils';
import type { DcaService } from './index';
import type { CreateDcaPlanRequest, DcaPlan, DcaRound, UpdateDcaPlanRequest } from '../../types';

const plans: DcaPlan[] = structuredClone(dcaData.plans) as DcaPlan[];
const rounds = structuredClone(dcaData.rounds) as Record<string, DcaRound[]>;
let planSeq = 4100;

function find(id: number): DcaPlan {
  const plan = plans.find((p) => p.id === id);
  if (!plan) throw new ApiClientError('VALIDATION_ERROR', 'DCA plan not found.', { status: 404 });
  return plan;
}

export const mockDcaService: DcaService = {
  async listPlans(): Promise<DcaPlan[]> {
    return mockResolve(plans);
  },

  async createPlan(req: CreateDcaPlanRequest): Promise<DcaPlan> {
    const min = req.assetType === 'gold' ? 100 : 500;
    if (Number(req.amountPerRound) < min) {
      throw new ApiClientError(
        'VALIDATION_ERROR',
        `Minimum per round is ${min} THB for ${req.assetType}.`,
      );
    }
    planSeq += 1;
    const now = new Date().toISOString();
    const plan: DcaPlan = {
      id: planSeq,
      userId: 1,
      assetType: req.assetType,
      assetId: req.assetId,
      amountPerRound: req.amountPerRound,
      frequency: req.frequency,
      runDay: req.runDay,
      startDate: req.startDate,
      endDate: req.endDate ?? null,
      maxRounds: req.maxRounds ?? null,
      roundsExecuted: 0,
      totalInvested: '0.0000',
      consecutiveFailures: 0,
      nextRunDate: req.startDate,
      status: 'active',
      createdAt: now,
      updatedAt: now,
    };
    plans.push(plan);
    rounds[String(plan.id)] = [];
    return mockResolve(plan);
  },

  async updatePlan(id: number, req: UpdateDcaPlanRequest): Promise<DcaPlan> {
    const plan = find(id);
    if (req.amountPerRound !== undefined) plan.amountPerRound = req.amountPerRound;
    if (req.runDay !== undefined) plan.runDay = req.runDay;
    plan.updatedAt = new Date().toISOString();
    return mockResolve(plan);
  },

  async pausePlan(id: number): Promise<DcaPlan> {
    const plan = find(id);
    plan.status = 'paused';
    plan.nextRunDate = null;
    plan.updatedAt = new Date().toISOString();
    return mockResolve(plan);
  },

  async resumePlan(id: number): Promise<DcaPlan> {
    const plan = find(id);
    plan.status = 'active';
    plan.updatedAt = new Date().toISOString();
    return mockResolve(plan);
  },

  async cancelPlan(id: number): Promise<void> {
    const plan = find(id);
    plan.status = 'cancelled';
    plan.nextRunDate = null;
    plan.updatedAt = new Date().toISOString();
    await mockResolve(null);
  },

  async listRounds(id: number): Promise<DcaRound[]> {
    return mockResolve(rounds[String(id)] ?? []);
  },
};
