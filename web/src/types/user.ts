// User, profile and risk-assessment DTOs (data-model.md: users, risk_assessments).

import type { ISODate, ISODateTime } from './common';

export type UserRole = 'investor' | 'admin';

export type UserStatus = 'pending_verification' | 'active' | 'locked';

/** Risk suitability level 1–5 (FR-04, BR-01). */
export type RiskLevel = 1 | 2 | 3 | 4 | 5;

export interface User {
  id: number;
  email: string;
  fullName: string;
  mobile: string;
  dateOfBirth: ISODate;
  role: UserRole;
  status: UserStatus;
  /** Masked on display (US-006) — e.g. "xxx-x-x1234-5". */
  bankAccountMasked: string | null;
  pdpaConsentAt: ISODateTime | null;
  emailNotificationsEnabled: boolean;
  createdAt: ISODateTime;
  updatedAt: ISODateTime;
}

/** Editable profile fields (PUT /api/profile). */
export interface ProfileUpdate {
  fullName: string;
  mobile: string;
  bankAccount: string;
}

export interface RiskAssessment {
  id: number;
  userId: number;
  riskLevel: RiskLevel;
  /** Questionnaire answers, keyed by question id. */
  answers: Record<string, number>;
  assessedAt: ISODateTime;
  /** assessedAt + 2 years (BR-01). */
  expiresAt: ISODateTime;
}

/** Editable profile as returned by GET /api/profile (bank account is masked). */
export interface Profile {
  id: number;
  email: string;
  fullName: string;
  mobile: string;
  bankAccountMasked: string;
  status: UserStatus;
}

/** PUT /api/profile body. Leave bankAccount out to keep the account on file. */
export interface UpdateProfileRequest {
  fullName: string;
  mobile: string;
  bankAccount?: string;
}

/** Latest risk assessment as returned by GET /api/risk-assessment. */
export interface RiskSummary {
  riskLevel: RiskLevel;
  assessedAt: ISODateTime;
  expiresAt: ISODateTime;
  /** False once the 2-year validity has passed (BR-01). */
  valid: boolean;
}
