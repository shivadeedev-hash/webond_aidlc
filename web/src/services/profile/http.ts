// HTTP profile implementation — identical interface to the mock.
import { apiClient } from '../../lib/apiClient';
import type { ProfileService } from './index';
import type {
  Profile,
  RiskLevel,
  RiskSummary,
  UpdateProfileRequest,
  UserStatus,
} from '../../types';

interface ApiProfile {
  id: number;
  email: string;
  fullName: string;
  mobile?: string;
  bankAccount?: string;
  status: UserStatus;
}

interface ApiRisk {
  riskLevel: RiskLevel;
  assessedAt: string;
  expiresAt: string;
  valid?: boolean;
}

function toProfile(p: ApiProfile): Profile {
  return {
    id: p.id,
    email: p.email,
    fullName: p.fullName,
    mobile: p.mobile ?? '',
    bankAccountMasked: p.bankAccount ?? '',
    status: p.status,
  };
}

function toRisk(r: ApiRisk): RiskSummary {
  return {
    riskLevel: r.riskLevel,
    assessedAt: r.assessedAt,
    expiresAt: r.expiresAt,
    valid: r.valid ?? true,
  };
}

export const httpProfileService: ProfileService = {
  async getProfile(): Promise<Profile> {
    const { data } = await apiClient.get<ApiProfile>('/profile');
    return toProfile(data);
  },

  async updateProfile(req: UpdateProfileRequest): Promise<Profile> {
    await apiClient.put('/profile', req);
    // PUT returns only the changed fields; re-read for the complete, masked profile.
    return this.getProfile();
  },

  async getRisk(): Promise<RiskSummary | null> {
    const { data } = await apiClient.get<ApiRisk | null>('/risk-assessment');
    return data ? toRisk(data) : null;
  },

  async submitRisk(answers: Record<string, number>, riskLevel: RiskLevel): Promise<RiskSummary> {
    const { data } = await apiClient.post<ApiRisk>('/risk-assessment', { answers, riskLevel });
    return toRisk(data);
  },
};
