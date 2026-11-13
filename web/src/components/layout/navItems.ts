// Primary navigation items (components.md AppShell). `protected` items require a
// signed-in investor; guests are redirected to Login (US-005).

export interface NavItem {
  path: string;
  /** i18n key for the label. */
  labelKey: string;
  protected: boolean;
}

export const NAV_ITEMS: readonly NavItem[] = [
  { path: '/dashboard', labelKey: 'nav.dashboard', protected: true },
  { path: '/market', labelKey: 'nav.market', protected: false },
  { path: '/portfolio', labelKey: 'nav.portfolio', protected: true },
  { path: '/dca', labelKey: 'nav.dca', protected: true },
  { path: '/history', labelKey: 'nav.history', protected: true },
  { path: '/profile', labelKey: 'nav.profile', protected: true },
] as const;
