// Risk-mismatch acknowledgement box (US-018, EX-12). Shown when the asset's risk
// level exceeds the investor's risk level; the order is blocked until checked.

import { useI18n } from '../../i18n/useI18n';
import { AlertTriangle } from '../ui/icons';

export default function RiskAckBox({
  assetRisk,
  userRisk,
  checked,
  onChange,
}: {
  assetRisk: number;
  userRisk: number;
  checked: boolean;
  onChange: (next: boolean) => void;
}): JSX.Element {
  const { t } = useI18n();
  return (
    <div className="rounded-md border border-warning bg-warning/10 p-3 text-sm">
      <p className="flex items-center gap-2 font-medium text-warning">
        <AlertTriangle size={16} />
        <span>
          Fund risk {assetRisk} &gt; your risk {userRisk}
        </span>
      </p>
      <label className="mt-2 flex items-center gap-2 text-text">
        <input
          type="checkbox"
          checked={checked}
          onChange={(e) => onChange(e.target.checked)}
          className="h-5 w-5 rounded border-border accent-[color:var(--color-primary)]"
        />
        <span>{t('page.detail.riskAck')}</span>
      </label>
    </div>
  );
}
