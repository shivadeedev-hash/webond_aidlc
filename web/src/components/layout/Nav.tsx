// Primary nav link list, reused for the desktop bar and the mobile drawer.
// Guests only see public items; protected links appear once authenticated (US-005).

import { NavLink } from 'react-router-dom';
import { NAV_ITEMS } from './navItems';
import { useAuth } from '../../hooks/useAuth';
import { useI18n } from '../../i18n/useI18n';

interface NavProps {
  orientation: 'horizontal' | 'vertical';
  onNavigate?: () => void;
}

export default function Nav({ orientation, onNavigate }: NavProps): JSX.Element {
  const { isAuthenticated } = useAuth();
  const { t } = useI18n();

  const items = NAV_ITEMS.filter((item) => !item.protected || isAuthenticated);

  const base = orientation === 'horizontal' ? 'flex items-center gap-1' : 'flex flex-col gap-1';

  return (
    <nav aria-label="Primary" className={base}>
      {items.map((item) => (
        <NavLink
          key={item.path}
          to={item.path}
          onClick={onNavigate}
          className={({ isActive }) =>
            [
              'rounded-full px-4 py-2 text-sm font-medium transition-standard duration-standard',
              'min-h-[44px] flex items-center',
              isActive
                ? 'bg-primary-soft font-semibold text-primary'
                : 'text-text-muted hover:bg-surface-muted hover:text-header-text',
            ].join(' ')
          }
        >
          {t(item.labelKey)}
        </NavLink>
      ))}
    </nav>
  );
}
