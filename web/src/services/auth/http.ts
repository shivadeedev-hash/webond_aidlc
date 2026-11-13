// HTTP auth implementation — identical interface to the mock. Real endpoints are
// exercised in the api wave; this wires the apiClient against api-spec.md.

import { apiClient, setAccessToken } from '../../lib/apiClient';
import type { AuthService } from './index';
import type {
  ForgotPasswordRequest,
  LoginRequest,
  LoginResponse,
  RefreshResponse,
  RegisterRequest,
  ResendVerificationRequest,
  ResetPasswordRequest,
  VerifyRequest,
} from '../../types';

export const httpAuthService: AuthService = {
  async register(req: RegisterRequest) {
    await apiClient.post('/auth/register', req);
  },

  async verify(req: VerifyRequest) {
    await apiClient.post('/auth/verify', req);
  },
  async resendVerification(req: ResendVerificationRequest) {
    await apiClient.post('/auth/resend-verification', req);
  },

  async login(req: LoginRequest): Promise<LoginResponse> {
    const { data } = await apiClient.post<LoginResponse>('/auth/login', req);
    setAccessToken(data.accessToken);
    return data;
  },

  async refresh(): Promise<RefreshResponse> {
    const { data } = await apiClient.post<RefreshResponse>('/auth/refresh');
    setAccessToken(data.accessToken);
    return data;
  },

  async logout() {
    await apiClient.post('/auth/logout');
    setAccessToken(null);
  },

  async forgotPassword(req: ForgotPasswordRequest) {
    await apiClient.post('/auth/forgot-password', req);
  },

  async resetPassword(req: ResetPasswordRequest) {
    await apiClient.post('/auth/reset-password', req);
  },
};
