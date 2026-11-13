// Auth service interface + env-selected implementation (api-spec.md /api/auth/*).

import { DATA_SOURCE } from '../../lib/constants';
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
import { mockAuthService } from './mock';
import { httpAuthService } from './http';

export interface AuthService {
  register(req: RegisterRequest): Promise<void>;
  verify(req: VerifyRequest): Promise<void>;
  resendVerification(req: ResendVerificationRequest): Promise<void>;
  login(req: LoginRequest): Promise<LoginResponse>;
  refresh(): Promise<RefreshResponse>;
  logout(): Promise<void>;
  forgotPassword(req: ForgotPasswordRequest): Promise<void>;
  resetPassword(req: ResetPasswordRequest): Promise<void>;
}

export const authService: AuthService = DATA_SOURCE === 'api' ? httpAuthService : mockAuthService;
