// Portfolio + dashboard service interface + env-selected implementation
// (api-spec.md /api/portfolio, /api/dashboard).

import { DATA_SOURCE } from '../../lib/constants';
import type { Dashboard, Portfolio, TrendPeriod } from '../../types';
import { mockPortfolioService } from './mock';
import { httpPortfolioService } from './http';

export interface PortfolioService {
  getPortfolio(): Promise<Portfolio>;
  getDashboard(trend?: TrendPeriod): Promise<Dashboard>;
}

export const portfolioService: PortfolioService =
  DATA_SOURCE === 'api' ? httpPortfolioService : mockPortfolioService;
