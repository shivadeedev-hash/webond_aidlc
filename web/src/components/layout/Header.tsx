// White sticky app header with red brand accents (Add-on 2026-11-12c): brand, primary nav,
// locale toggle, and the auth action. Collapses to a hamburger + drawer on mobile.

import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import Container from './Container';
import Nav from './Nav';
import Logo from './Logo';
import LocaleToggle from './LocaleToggle';
import { useAuth } from '../../hooks/useAuth';
import { useI18n } from '../../i18n/useI18n';

export default function Header(): JSX.Element {
  const { isAuthenticated, logout } = useAuth();
  const { t } = useI18n();
  const navigate = useNavigate();
  const [menuOpen, setMenuOpen] = useState(false);

  async function handleLogout(): Promise<void> {
    await logout();
    setMenuOpen(false);
    navigate('/market');
  }

  const authAction = isAuthenticated ? (
    <button
      type="button"
      onClick={handleLogout}
      className="min-h-[44px] rounded-full border border-border px-5 text-sm font-medium text-header-text transition-standard duration-standard hover:border-primary hover:text-primary"
    >
      {t('nav.logout')}
    </button>
  ) : (
    <Link
      to="/login"
      onClick={() => setMenuOpen(false)}
      className="flex min-h-[44px] items-center rounded-full bg-primary px-5 text-sm font-semibold text-text-inverse shadow-sm transition-standard duration-standard hover:bg-primary-hover active:bg-primary-active"
    >
      {t('nav.login')}
    </Link>
  );

  return (
    <header className="sticky top-0 z-30 border-b border-border bg-header/90 text-header-text shadow-sm backdrop-blur supports-[backdrop-filter]:bg-header/80">
      <Container className="flex items-center justify-between gap-4 py-3">
        <Link to="/" className="flex items-center" aria-label={t('app.name')}>
          <Logo />
        </Link>

        {/* Desktop nav */}
        <div className="hidden items-center gap-4 lg:flex">
          <Nav orientation="horizontal" />
          <LocaleToggle />
          {authAction}
        </div>

        {/* Mobile menu toggle */}
        <button
          type="button"
          className="flex h-11 w-11 items-center justify-center rounded-full border border-border text-header-text hover:border-primary hover:text-primary lg:hidden"
          aria-expanded={menuOpen}
          aria-controls="mobile-menu"
          aria-label={menuOpen ? t('nav.closeMenu') : t('nav.openMenu')}
          onClick={() => setMenuOpen((o) => !o)}
        >
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            {menuOpen ? (
              <path
                d="M6 6l12 12M18 6L6 18"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
              />
            ) : (
              <path
                d="M4 7h16M4 12h16M4 17h16"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
              />
            )}
          </svg>
        </button>
      </Container>

      {/* Mobile drawer */}
      {menuOpen && (
        <div id="mobile-menu" className="border-t border-border bg-header lg:hidden">
          <Container className="flex flex-col gap-4 py-4">
            <Nav orientation="vertical" onNavigate={() => setMenuOpen(false)} />
            <div className="flex items-center justify-between gap-3">
              <LocaleToggle />
              {authAction}
            </div>
          </Container>
        </div>
      )}
    </header>
  );
}
