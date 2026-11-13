// Profile + risk-assessment service interface and env-selected implementation
// (api-spec.md /api/profile, /api/risk-assessment).
import { DATA_SOURCE } from '../../lib/constants';
import type { Profile, RiskLevel, RiskSummary, UpdateProfileRequest } from '../../types';
import { mockProfileService } from './mock';
import { httpProfileService } from './http';

export interface ProfileService {
  getProfile(): Promise<Profile>;
  updateProfile(req: UpdateProfileRequest): Promise<Profile>;
  /** The latest assessment, or null when the user has never taken one. */
  getRisk(): Promise<RiskSummary | null>;
  submitRisk(answers: Record<string, number>, riskLevel: RiskLevel): Promise<RiskSummary>;
}

export const profileService: ProfileService =
  DATA_SOURCE === 'api' ? httpProfileService : mockProfileService;
