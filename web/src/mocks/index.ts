// Typed access to the mock JSON seed (US-032). Shapes mirror the API DTOs
// described in data-model.md / api-spec.md so mock.ts and http.ts stay identical.
import fundsJson from './funds.json';
import bondsJson from './bonds.json';
import goldJson from './gold.json';
import holdingsJson from './holdings.json';
import dcaJson from './dca.json';
import transactionsJson from './transactions.json';
import notificationsJson from './notifications.json';
import walletJson from './wallet.json';
import userJson from './user.json';

import type {
  Bond,
  DcaPlan,
  Fund,
  Gold,
  Holding,
  Notification,
  RiskAssessment,
  Transaction,
  User,
  Wallet,
  WalletTransaction,
} from '../types';

export const mockFunds = fundsJson as unknown as Fund[];
export const mockBonds = bondsJson as unknown as Bond[];
export const mockGold = goldJson as unknown as Gold[];
export const mockHoldings = holdingsJson as unknown as Holding[];
export const mockDcaPlans = dcaJson.plans as unknown as DcaPlan[];
export const mockTransactions = transactionsJson.transactions as unknown as Transaction[];
export const mockNotifications = notificationsJson.notifications as unknown as Notification[];

export const mockWallet = walletJson.wallet as unknown as Wallet;
export const mockWalletTransactions = walletJson.transactions as unknown as WalletTransaction[];

export const mockUser = userJson.user as unknown as User;
export const mockRiskAssessment = userJson.riskAssessment as unknown as RiskAssessment;
export const demoCredentials = userJson.demoCredentials as { email: string; password: string };
