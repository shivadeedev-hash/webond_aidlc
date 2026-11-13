// DCA create / edit form (US-019, wireframe 05). Mutual Fund & Gold only —
// bonds are excluded (BR-06). Enforces min amount 500 MF / 100 Gold (BR-07)
// client-side; the service re-validates. Edit mode locks the asset and only
// allows amount + run day changes (matches UpdateDcaPlanRequest).

import { useMemo, useState, type FormEvent } from 'react';
import type {
  AssetListItem,
  CreateDcaPlanRequest,
  DcaAssetType,
  DcaFrequency,
  DcaPlan,
  UpdateDcaPlanRequest,
} from '../../types';
import { useI18n } from '../../i18n/useI18n';
import { ApiClientError } from '../../lib/errors';
import Button from '../ui/Button';
import { TextField, SelectField } from '../ui/Field';
import AmountInput from '../trade/AmountInput';
import { amountValue, floorTo, sanitizeAmount } from '../../lib/amountInput';

interface Props {
  /** Fund + gold assets available to pick from. */
  assets: AssetListItem[];
  /** Plan being edited, or null for a create. */
  editing: DcaPlan | null;
  /** Preselected asset (from the "Set up DCA" shortcut). */
  preset?: { assetType: DcaAssetType; assetId: number } | null;
  onCreate: (req: CreateDcaPlanRequest) => Promise<void>;
  onUpdate: (id: number, req: UpdateDcaPlanRequest) => Promise<void>;
  onCancel: () => void;
}

function minFor(type: DcaAssetType): number {
  return type === 'gold' ? 100 : 500;
}

/** Quick amounts offered for a plan: the minimum plus a few common steps. */
function picksFor(type: DcaAssetType): number[] {
  const min = minFor(type);
  return [min, 1000, 5000, 10000].filter((v, i, a) => v >= min && a.indexOf(v) === i);
}

export default function PlanForm({
  assets,
  editing,
  preset,
  onCreate,
  onUpdate,
  onCancel,
}: Props): JSX.Element {
  const { t } = useI18n();

  const initialAssetKey = editing
    ? `${editing.assetType}:${editing.assetId}`
    : preset
      ? `${preset.assetType}:${preset.assetId}`
      : assets[0]
        ? `${assets[0].assetType}:${assets[0].id}`
        : '';

  const [assetKey, setAssetKey] = useState(initialAssetKey);
  // Plain numeric string ("5000"); the field shows it with thousands separators.
  const [amount, setAmount] = useState(editing ? floorTo(Number(editing.amountPerRound), 2) : '');
  const [frequency, setFrequency] = useState<DcaFrequency>(editing?.frequency ?? 'monthly');
  const [runDay, setRunDay] = useState(String(editing?.runDay ?? 1));
  const [startDate, setStartDate] = useState(
    editing?.startDate ?? new Date().toISOString().slice(0, 10),
  );
  const [endDate, setEndDate] = useState(editing?.endDate ?? '');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const selected = useMemo(
    () => assets.find((a) => `${a.assetType}:${a.id}` === assetKey),
    [assets, assetKey],
  );
  const assetType: DcaAssetType = selected?.assetType === 'gold' ? 'gold' : 'fund';

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    setError(null);

    const amountNum = amountValue(sanitizeAmount(amount, 2));
    const min = minFor(assetType);
    if (!(amountNum >= min)) {
      setError(
        `Minimum per round is ${min} THB for ${assetType === 'gold' ? 'Gold' : 'Mutual Fund'} (BR-07).`,
      );
      return;
    }

    setSubmitting(true);
    try {
      if (editing) {
        await onUpdate(editing.id, { amountPerRound: String(amountNum), runDay: Number(runDay) });
      } else {
        if (!selected) {
          setError(t('errors.VALIDATION_ERROR'));
          return;
        }
        await onCreate({
          assetType,
          assetId: selected.id,
          amountPerRound: String(amountNum),
          frequency,
          runDay: Number(runDay),
          startDate,
          endDate: endDate || undefined,
        });
      }
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : t('errors.UNKNOWN'));
    } finally {
      setSubmitting(false);
    }
  }

  const runDayMax = frequency === 'weekly' ? 7 : 28;

  return (
    <form className="flex flex-col gap-5" onSubmit={handleSubmit} noValidate>
      <h3 className="text-base font-semibold">{t('page.dca.editTitle')}</h3>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3 md:items-start">
        <SelectField
          label={t('page.dca.col.asset')}
          value={assetKey}
          onChange={(e) => setAssetKey(e.target.value)}
          disabled={editing !== null}
        >
          {assets.map((a) => (
            <option key={`${a.assetType}:${a.id}`} value={`${a.assetType}:${a.id}`}>
              {a.code} — {a.assetType === 'gold' ? 'Gold' : 'Fund'}
            </option>
          ))}
        </SelectField>

        <AmountInput
          label={t('page.dca.amountPerRound')}
          unit={t('trade.baht')}
          value={amount}
          decimals={2}
          onChange={setAmount}
          picks={picksFor(assetType).map((v) => ({
            label: String(v.toLocaleString('en-US')),
            value: String(v),
          }))}
          picksLabel={t('trade.quickPicks')}
          hint={`Min ${minFor(assetType)} THB`}
        />

        <SelectField
          label={t('page.dca.frequency')}
          value={frequency}
          onChange={(e) => setFrequency(e.target.value as DcaFrequency)}
          disabled={editing !== null}
        >
          <option value="weekly">Weekly</option>
          <option value="monthly">Monthly</option>
        </SelectField>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <TextField
          label={`${t('page.dca.runDay')} (1–${runDayMax})`}
          type="number"
          min="1"
          max={runDayMax}
          value={runDay}
          onChange={(e) => setRunDay(e.target.value)}
        />
        <TextField
          label={t('page.dca.startDate')}
          type="date"
          value={startDate}
          onChange={(e) => setStartDate(e.target.value)}
          disabled={editing !== null}
        />
        <TextField
          label={t('page.dca.endDate')}
          type="date"
          value={endDate}
          onChange={(e) => setEndDate(e.target.value)}
          disabled={editing !== null}
        />
      </div>

      {error && (
        <p role="alert" className="text-sm text-danger">
          {error}
        </p>
      )}

      <div className="flex justify-end gap-3">
        <Button variant="outline" type="button" onClick={onCancel}>
          {t('common.cancel')}
        </Button>
        <Button variant="primary" type="submit" disabled={submitting}>
          {submitting ? t('common.loading') : t('page.dca.savePlan')}
        </Button>
      </div>
    </form>
  );
}
