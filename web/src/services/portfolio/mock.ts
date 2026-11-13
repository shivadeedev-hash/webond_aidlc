// Mock portfolio + dashboard implementation backed by holdings.json / dashboard.json.

import holdingsData from '../../mocks/holdings.json';
import dashboardData from '../../mocks/dashboard.json';
import { mockResolve } from '../mockUtils';
import type { PortfolioService } from './index';
import type { Dashboard, Holding, Portfolio, PortfolioSummary } from '../../types';

const holdings = holdingsData as Holding[];

function summarize(items: Holding[]): PortfolioSummary {
  const totalCost = items.reduce((s, h) => s + Number(h.totalCost), 0);
  const totalMarketValue = items.reduce((s, h) => s + Number(h.marketValue), 0);
  const totalUnrealizedPl = totalMarketValue - totalCost;
  const pct = totalCost > 0 ? (totalUnrealizedPl / totalCost) * 100 : 0;
  return {
    totalCost: totalCost.toFixed(2),
    totalMarketValue: totalMarketValue.toFixed(2),
    totalUnrealizedPl: totalUnrealizedPl.toFixed(2),
    totalUnrealizedPlPct: pct.toFixed(2),
  };
}

export const mockPortfolioService: PortfolioService = {
  async getPortfolio(): Promise<Portfolio> {
    return mockResolve<Portfolio>({ summary: summarize(holdings), holdings });
  },

  async getDashboard(): Promise<Dashboard> {
    return mockResolve(dashboardData as Dashboard);
  },
};
