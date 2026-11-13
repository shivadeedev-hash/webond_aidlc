// Auth context object + value type, kept separate from the provider component so
// the hook and provider can each live in a component-only / hook-only module
// (satisfies react-refresh/only-export-components).

import { createContext } from 'react';
import type { LoginRequest, User } from '../types';

export interface AuthContextValue {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  /** True while a saved session is being restored after a page load. */
  isInitializing: boolean;
  login: (req: LoginRequest) => Promise<User>;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextValue | null>(null);
