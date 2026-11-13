// HTTP DCA implementation — same interface as the mock, API shapes mapped
// through apiMappers.
import { apiClient } from '../../lib/apiClient';
import type { DcaService } from './index';
import type { CreateDcaPlanRequest, DcaPlan, DcaRound, UpdateDcaPlanRequest } from '../../types';
import { toDcaPlan, toDcaRound, type ApiDcaPlan, type ApiDcaRound } from '../apiMappers';

/** Pause/resume answer only { id, status }; re-read the plan for the full record. */
async function reload(id: number): Promise<DcaPlan> {
  const { data } = await apiClient.get<ApiDcaPlan[]>('/dca-plans');
  const plan = (data ?? []).find((p) => p.id === id);
  if (!plan) throw new Error(`DCA plan ${id} not found after update`);
  return toDcaPlan(plan);
}

export const httpDcaService: DcaService = {
  async listPlans(): Promise<DcaPlan[]> {
    const { data } = await apiClient.get<ApiDcaPlan[]>('/dca-plans');
    return (data ?? []).map(toDcaPlan);
  },
  async createPlan(req: CreateDcaPlanRequest): Promise<DcaPlan> {
    const { data } = await apiClient.post<ApiDcaPlan>('/dca-plans', req);
    return toDcaPlan(data);
  },
  async updatePlan(id: number, req: UpdateDcaPlanRequest): Promise<DcaPlan> {
    const { data } = await apiClient.put<ApiDcaPlan>(`/dca-plans/${id}`, req);
    return toDcaPlan(data);
  },
  async pausePlan(id: number): Promise<DcaPlan> {
    await apiClient.post(`/dca-plans/${id}/pause`);
    return reload(id);
  },
  async resumePlan(id: number): Promise<DcaPlan> {
    await apiClient.post(`/dca-plans/${id}/resume`);
    return reload(id);
  },
  async cancelPlan(id: number): Promise<void> {
    await apiClient.delete(`/dca-plans/${id}`);
  },
  async listRounds(id: number): Promise<DcaRound[]> {
    const { data } = await apiClient.get<ApiDcaRound[]>(`/dca-plans/${id}/rounds`);
    return (data ?? []).map((r) => toDcaRound(r, id));
  },
};
