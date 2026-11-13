// React provider for the i18n scaffold. Exposes the current locale, a setter,
// and a t() helper. Thai is the default (NFR-07).

import { useCallback, useMemo, useState, type ReactNode } from 'react';
import { DEFAULT_LOCALE, SUPPORTED_LOCALES, translate, type Locale } from '../lib/i18n';
import { I18nContext, type I18nContextValue } from './i18nContext';

export function I18nProvider({
  children,
  initialLocale = DEFAULT_LOCALE,
}: {
  children: ReactNode;
  initialLocale?: Locale;
}): JSX.Element {
  const [locale, setLocale] = useState<Locale>(initialLocale);

  const t = useCallback((key: string) => translate(locale, key), [locale]);

  const value = useMemo<I18nContextValue>(
    () => ({ locale, setLocale, t, locales: SUPPORTED_LOCALES }),
    [locale, t],
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}
