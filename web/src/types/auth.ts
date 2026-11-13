// Auth request/response DTOs (api-spec.md: /api/auth/*).

import type { ISODate } from './common';
import type { User } from './user';

export interface RegisterRequest {
  email: string;
  password: string;
  fullName: string;
  mobile: string;
  dateOfBirth: ISODate;
  /** PDPA consent acknowledgement (required). */
  consent: boolean;
}

export interface LoginRequest {
  email: string;
  password: string;
}

/** Login response: access token (kept in memory) + the authenticated user.
 *  The refresh token is delivered via an httpOnly cookie, never in the body. */
export interface LoginResponse {
  accessToken: string;
  user: User;
}

/** Refresh response: a rotated access token, plus the user so the SPA can
 *  restore its session after a page reload (omitted by the mock). */
export interface RefreshResponse {
  accessToken: string;
  user?: User;
}

/** Verify an account with the 6-digit code from the e-mail (US-002). */
export interface VerifyRequest {
  email: string;
  code: string;
}

export interface ResendVerificationRequest {
  email: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

/** Set a new password with the 6-digit code from the e-mail (US-004). */
export interface ResetPasswordRequest {
  email: string;
  code: string;
  newPassword: string;
}
