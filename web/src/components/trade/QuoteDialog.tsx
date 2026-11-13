// Gold 30-second quote confirmation dialog (US-017, EX-08). Shows the locked
// price/weight/total with a live countdown; on expiry the confirm button is
// replaced by a re-quote action (QUOTE_EXPIRED handling).

import { useEffect, useState } from 'react';
import type { GoldQuote, OrderSide } from '../../types';
import { formatGrams, formatTHB } from '../../lib/format';
import { useI18n } from '../../i18n/useI18n';
import Button from '../ui/Button';
import { Clock } from '../ui/icons';

export default function QuoteDialog({
  quote,
  side,
  onConfirm,
  onRequote,
  onClose,
  submitting,
  error,
}: {
  quote: GoldQuote;
  side: OrderSide;
  onConfirm: () => void;
  onRequote: () => void;
  onClose: () => void;
  submitting: boolean;
  error: string | null;
}): JSX.Element {
  const { t } = useI18n();
  const [secondsLeft, setSecondsLeft] = useState(() =>
    Math.max(0, Math.ceil((new Date(quote.expiresAt).getTime() - Date.now()) / 1000)),
  );

  useEffect(() => {
    const expiry = new Date(quote.expiresAt).getTime();
    setSecondsLeft(Math.max(0, Math.ceil((expiry - Date.now()) / 1000)));
    const id = window.setInterval(() => {
      const left = Math.max(0, Math.ceil((expiry - Date.now()) / 1000));
      setSecondsLeft(left);
      if (left <= 0) window.clearInterval(id);
    }, 250);
    return () => window.clearInterval(id);
  }, [quote.expiresAt, quote.quoteId]);

  const expired = secondsLeft <= 0;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={side === 'buy' ? t('page.detail.buy') : t('page.detail.sell')}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-sm rounded-lg border border-border bg-surface p-6 shadow-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold">
          {side === 'buy' ? t('page.detail.buy') : t('page.detail.sell')} ·{' '}
          {t('page.market.tab.gold')}
        </h2>

        <dl className="mt-4 space-y-2 text-sm">
          <div className="flex justify-between">
            <dt className="text-text-muted">{side === 'buy' ? 'ASK/g' : 'BID/g'}</dt>
            <dd className="font-medium">{formatTHB(quote.price)}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-text-muted">{t('page.detail.weight')}</dt>
            <dd className="font-medium">{formatGrams(quote.weight)} g</dd>
          </div>
          <div className="flex justify-between border-t border-border pt-2">
            <dt className="text-text-muted">Total</dt>
            <dd className="font-semibold">{formatTHB(quote.total)}</dd>
          </div>
        </dl>

        <p
          className={`mt-4 inline-flex items-center gap-1.5 text-sm font-medium ${expired ? 'text-danger' : 'text-warning'}`}
          aria-live="polite"
        >
          <Clock size={16} />
          {expired
            ? t('errors.QUOTE_EXPIRED')
            : `${t('page.detail.quoteExpiresIn')} ${secondsLeft}s`}
        </p>

        {error && (
          <p role="alert" className="mt-2 text-sm text-danger">
            {error}
          </p>
        )}

        <div className="mt-5 flex gap-3">
          <Button variant="outline" className="flex-1" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          {expired ? (
            <Button variant="primary" className="flex-1" onClick={onRequote}>
              {t('page.detail.getQuote')}
            </Button>
          ) : (
            <Button variant="primary" className="flex-1" onClick={onConfirm} disabled={submitting}>
              {submitting ? t('common.loading') : t('page.detail.confirmQuote')}
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
