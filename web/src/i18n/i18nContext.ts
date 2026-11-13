// i18n context object + value type, separated from the provider component so the
// hook and provider live in component-only / hook-only modules
// (satisfies react-refresh/only-export-components).

import { createContext } from 'react';
import type { Locale } from '../lib/i18n';

export interface I18nContextValue {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: string) => string;
  locales: readonly Locale[];
}

export const I18nContext = createContext<I18nContextValue | null>(null);
