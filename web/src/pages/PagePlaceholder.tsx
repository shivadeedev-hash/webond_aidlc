// Shared placeholder for routed pages whose real content lands in later waves.
// Renders a titled, token-styled card so navigation and layout are verifiable now.

import { useI18n } from '../i18n/useI18n';

export default function PagePlaceholder({
  title,
  description,
}: {
  title: string;
  description?: string;
}): JSX.Element {
  const { t } = useI18n();
  return (
    <section className="rounded-lg border border-border bg-surface p-8 shadow-sm">
      <h1 className="text-2xl font-semibold">{title}</h1>
      <p className="mt-2 text-text-muted">{description ?? t('page.comingSoon')}</p>
    </section>
  );
}
