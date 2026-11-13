// Mock auth implementation backed by mocks/user.json. Accepts the printed demo
// credentials (US-032); other credentials fail with UNAUTHORIZED.

import userData from '../../mocks/user.json';
import { ApiClientError } from '../../lib/errors';
import { mockResolve } from '../mockUtils';
import type { AuthService } from './index';
import type { LoginRequest, LoginResponse, User } from '../../types';

const demo = userData.demoCredentials;
const demoUser = userData.user as User;

export const mockAuthService: AuthService = {
  async register() {
    await mockResolve(null);
  },

  async verify(req) {
    // Mock mode has no e-mail: any 6-digit code is accepted.
    if (!/^\d{6}$/.test(req.code)) {
      throw new ApiClientError('VALIDATION_ERROR', 'Enter the 6-digit code from the e-mail');
    }
    await mockResolve(null);
  },

  async resendVerification() {
    await mockResolve(null);
  },

  async login(req: LoginRequest): Promise<LoginResponse> {
    if (req.email !== demo.email || req.password !== demo.password) {
      throw new ApiClientError('UNAUTHORIZED', 'Invalid email or password.');
    }
    return mockResolve<LoginResponse>({
      accessToken: `mock-access-token.${Date.now()}`,
      user: demoUser,
    });
  },

  async refresh() {
    return mockResolve({ accessToken: `mock-access-token.${Date.now()}` });
  },

  async logout() {
    await mockResolve(null);
  },

  async forgotPassword() {
    // Always succeeds (no account enumeration) — mirrors the 202 contract.
    await mockResolve(null);
  },

  async resetPassword(req) {
    if (!/^\d{6}$/.test(req.code)) {
      throw new ApiClientError('VALIDATION_ERROR', 'Enter the 6-digit code from the e-mail');
    }
    await mockResolve(null);
  },
};
