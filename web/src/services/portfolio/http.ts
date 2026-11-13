// HTTP portfolio + dashboard implementation — same interface as the mock, API
// shapes mapped through apiMappers.
import { apiClient } from '../../lib/apiClient';
import type { PortfolioService } from './index';
import type { Dashboard, Portfolio, TrendPeriod } from '../../types';
import { toDashboard, toPortfolio, type ApiDashboard, type ApiPortfolio } from '../apiMappers';

export const httpPortfolioService: PortfolioService = {
  async getPortfolio(): Promise<Portfolio> {
    const { data } = await apiClient.get<ApiPortfolio>('/portfolio');
    return toPortfolio(data);
  },
  async getDashboard(trend: TrendPeriod = '1Y'): Promise<Dashboard> {
    const { data } = await apiClient.get<ApiDashboard>('/dashboard', { trend });
    return toDashboard(data);
  },
};
