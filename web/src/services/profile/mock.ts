// Mock profile implementation backed by mocks/user.json. State lives in memory so
// edits and a retaken assessment are visible for the rest of the session.
import userData from '../../mocks/user.json';
import { ApiClientError } from '../../lib/errors';
import { mockResolve } from '../mockUtils';
import type { ProfileService } from './index';
import type { Profile, RiskLevel, RiskSummary, UserStatus } from '../../types';

const TWO_YEARS_MS = 2 * 365 * 24 * 60 * 60 * 1000;

let profile: Profile = {
  id: userData.user.id,
  email: userData.user.email,
  fullName: userData.user.fullName,
  mobile: userData.user.mobile,
  bankAccountMasked: userData.user.bankAccountMasked ?? '',
  status: userData.user.status as UserStatus,
};

let risk: RiskSummary | null = {
  riskLevel: userData.riskAssessment.riskLevel as RiskLevel,
  assessedAt: userData.riskAssessment.assessedAt,
  expiresAt: userData.riskAssessment.expiresAt,
  valid: true,
};

/** Masks all but the last 4 characters, like the API does. */
function mask(account: string): string {
  const trimmed = account.trim();
  if (trimmed.length <= 4) return 'x'.repeat(trimmed.length);
  return 'x'.repeat(trimmed.length - 4) + trimmed.slice(-4);
}

export const mockProfileService: ProfileService = {
  async getProfile(): Promise<Profile> {
    return mockResolve({ ...profile });
  },

  async updateProfile(req): Promise<Profile> {
    if (req.fullName.trim() === '') {
      throw new ApiClientError('VALIDATION_ERROR', 'Full name is required');
    }
    profile = {
      ...profile,
      fullName: req.fullName.trim(),
      mobile: req.mobile.trim(),
      // Blank means "unchanged", exactly like the API.
      bankAccountMasked: req.bankAccount?.trim()
        ? mask(req.bankAccount)
        : profile.bankAccountMasked,
    };
    return mockResolve({ ...profile });
  },

  async getRisk(): Promise<RiskSummary | null> {
    return mockResolve(risk ? { ...risk } : null);
  },

  async submitRisk(_answers, riskLevel): Promise<RiskSummary> {
    const now = new Date();
    risk = {
      riskLevel,
      assessedAt: now.toISOString(),
      expiresAt: new Date(now.getTime() + TWO_YEARS_MS).toISOString(),
      valid: true,
    };
    return mockResolve({ ...risk });
  },
};
