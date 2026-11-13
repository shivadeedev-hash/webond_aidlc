// Barrel for all WeBond domain services. Each domain selects its implementation
// (mock JSON | HTTP) via VITE_DATA_SOURCE at import time (see lib/constants.ts).

export { authService, type AuthService } from './auth';
export { catalogService, type CatalogService, type AssetQueryParams } from './catalog';
export { walletService, type WalletService } from './wallet';
export { orderService, type OrderService } from './order';
export { dcaService, type DcaService } from './dca';
export { portfolioService, type PortfolioService } from './portfolio';
export { historyService, type HistoryService } from './history';
export { notificationService, type NotificationService } from './notification';
export { profileService, type ProfileService } from './profile';
