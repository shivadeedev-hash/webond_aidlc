// Deposit / Withdraw dialog (US-008/009). Withdraw is capped at available balance
// (surfaces INSUFFICIENT_BALANCE from the service). Deposit is mocked payment.

import { useState, type FormEvent } from 'react';
import { walletService } from '../../services/wallet';
import { ApiClientError } from '../../lib/errors';
import { formatTHB } from '../../lib/format';
import { useI18n } from '../../i18n/useI18n';
import Button from '../ui/Button';
import { TextField } from '../ui/Field';

export default function WalletActionDialog({
  mode,
  available,
  onClose,
  onSuccess,
}: {
  mode: 'deposit' | 'withdraw';
  available: number;
  onClose: () => void;
  onSuccess: () => void;
}): JSX.Element {
  const { t } = useI18n();
  const [amount, setAmount] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    setError(null);
    const n = Number(amount);
    if (!(n > 0)) {
      setError(t('errors.VALIDATION_ERROR'));
      return;
    }
    setSubmitting(true);
    try {
      if (mode === 'deposit') await walletService.deposit({ amount });
      else await walletService.withdraw({ amount });
      onSuccess();
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : t('errors.UNKNOWN'));
    } finally {
      setSubmitting(false);
    }
  }

  const title = mode === 'deposit' ? t('page.profile.deposit') : t('page.profile.withdraw');

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={title}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-sm rounded-lg border border-border bg-surface p-6 shadow-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold">{title}</h2>
        {mode === 'withdraw' && (
          <p className="mt-1 text-sm text-text-muted">
            {t('page.profile.available')}: {formatTHB(available)}
          </p>
        )}
        <form className="mt-4 flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
          <TextField
            label={t('page.detail.amount')}
            type="number"
            inputMode="decimal"
            min="0"
            step="any"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            error={error ?? undefined}
            autoFocus
          />
          <div className="flex gap-3">
            <Button variant="outline" className="flex-1" onClick={onClose} type="button">
              {t('common.cancel')}
            </Button>
            <Button variant="primary" className="flex-1" type="submit" disabled={submitting}>
              {submitting ? t('common.loading') : t('common.confirm')}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
