// DCA plan table (US-020, wireframe 05). Shows asset, type, amount per round,
// frequency, next run, invested-to-date, status, and manage actions
// (edit / pause / resume / cancel). Status pairs color + label (not color alone).

import type { DcaPlan } from '../../types';
import { formatDateBangkok, formatTHB } from '../../lib/format';
import { useI18n } from '../../i18n/useI18n';

function statusTone(status: DcaPlan['status']): string {
  switch (status) {
    case 'active':
      return 'text-success';
    case 'paused':
      return 'text-warning';
    case 'completed':
      return 'text-info';
    case 'cancelled':
      return 'text-text-muted';
  }
}

function frequencyLabel(plan: DcaPlan): string {
  if (plan.frequency === 'monthly') return `Monthly, day ${plan.runDay}`;
  const days = ['', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
  return `Weekly, ${days[plan.runDay] ?? plan.runDay}`;
}

export default function PlanTable({
  plans,
  codeFor,
  onEdit,
  onPause,
  onResume,
  onCancel,
  busyId,
}: {
  plans: DcaPlan[];
  codeFor: (assetType: string, assetId: number) => string;
  onEdit: (plan: DcaPlan) => void;
  onPause: (id: number) => void;
  onResume: (id: number) => void;
  onCancel: (id: number) => void;
  busyId: number | null;
}): JSX.Element {
  const { t } = useI18n();

  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[820px] text-sm">
        <thead>
          <tr className="border-b border-border bg-surface-muted text-left text-text-muted">
            <th scope="col" className="px-4 py-3 font-medium">
              {t('page.dca.col.asset')}
            </th>
            <th scope="col" className="px-4 py-3 font-medium">
              {t('page.dca.col.type')}
            </th>
            <th scope="col" className="px-4 py-3 text-right font-medium">
              {t('page.dca.col.amount')}
            </th>
            <th scope="col" className="px-4 py-3 font-medium">
              {t('page.dca.col.frequency')}
            </th>
            <th scope="col" className="px-4 py-3 font-medium">
              {t('page.dca.col.nextRun')}
            </th>
            <th scope="col" className="px-4 py-3 text-right font-medium">
              {t('page.dca.col.invested')}
            </th>
            <th scope="col" className="px-4 py-3 font-medium">
              {t('page.dca.col.status')}
            </th>
            <th scope="col" className="px-4 py-3 font-medium">
              {t('page.dca.col.actions')}
            </th>
          </tr>
        </thead>
        <tbody>
          {plans.map((plan) => {
            const terminal = plan.status === 'cancelled' || plan.status === 'completed';
            return (
              <tr key={plan.id} className="border-b border-border last:border-0 align-middle">
                <td className="px-4 py-3 font-medium">{codeFor(plan.assetType, plan.assetId)}</td>
                <td className="px-4 py-3">{plan.assetType === 'gold' ? 'Gold' : 'Fund'}</td>
                <td className="px-4 py-3 text-right">{formatTHB(plan.amountPerRound)}</td>
                <td className="px-4 py-3 whitespace-nowrap">{frequencyLabel(plan)}</td>
                <td className="px-4 py-3 whitespace-nowrap">
                  {plan.nextRunDate ? formatDateBangkok(plan.nextRunDate) : '—'}
                </td>
                <td className="px-4 py-3 text-right">{formatTHB(plan.totalInvested)}</td>
                <td className={`px-4 py-3 font-medium capitalize ${statusTone(plan.status)}`}>
                  {plan.status}
                </td>
                <td className="px-4 py-3">
                  <div className="flex flex-wrap gap-2">
                    {!terminal && (
                      <>
                        <button
                          type="button"
                          onClick={() => onEdit(plan)}
                          className="text-primary hover:underline disabled:opacity-50"
                          disabled={busyId === plan.id}
                        >
                          Edit
                        </button>
                        {plan.status === 'active' ? (
                          <button
                            type="button"
                            onClick={() => onPause(plan.id)}
                            className="text-warning hover:underline disabled:opacity-50"
                            disabled={busyId === plan.id}
                          >
                            {t('page.dca.pause')}
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => onResume(plan.id)}
                            className="text-success hover:underline disabled:opacity-50"
                            disabled={busyId === plan.id}
                          >
                            {t('page.dca.resume')}
                          </button>
                        )}
                        <button
                          type="button"
                          onClick={() => onCancel(plan.id)}
                          className="text-danger hover:underline disabled:opacity-50"
                          disabled={busyId === plan.id}
                        >
                          {t('common.cancel')}
                        </button>
                      </>
                    )}
                    {terminal && <span className="text-text-muted">—</span>}
                  </div>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
