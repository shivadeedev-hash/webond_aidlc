// Holdings table (US-022, wireframe 06). Columns: asset, type, qty, avg cost,
// market value, unrealized P/L (color + icon + sign), next event.
// Qty unit follows asset type: units (fund), units (bond), grams (gold).

import { Link } from 'react-router-dom';
import type { AssetType, Holding } from '../../types';
import { formatDateBangkok, formatMoney, formatTHB } from '../../lib/format';
import { useI18n } from '../../i18n/useI18n';
import PLValue from '../ui/PLValue';

function typeLabel(type: AssetType): string {
  if (type === 'fund') return 'Fund';
  if (type === 'bond') return 'Bond';
  return 'Gold';
}

function qtyDisplay(h: Holding): string {
  const q = formatMoney(h.quantity); // 2dp grouping is fine for display; units/grams noted by suffix
  if (h.assetType === 'gold') return `${q} g`;
  if (h.assetType === 'bond') return `${q} units`;
  return `${q} u`;
}

export default function HoldingsTable({ holdings }: { holdings: Holding[] }): JSX.Element {
  const { t } = useI18n();

  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[720px] text-sm">
        <thead>
          <tr className="border-b border-border bg-surface-muted text-left text-text-muted">
            <th scope="col" className="px-4 py-3 font-medium">
              {t('page.dca.col.asset')}
            </th>
            <th scope="col" className="px-4 py-3 font-medium">
              {t('page.dca.col.type')}
            </th>
            <th scope="col" className="px-4 py-3 text-right font-medium">
              {t('page.portfolio.col.qty')}
            </th>
            <th scope="col" className="px-4 py-3 text-right font-medium">
              {t('page.portfolio.col.avgCost')}
            </th>
            <th scope="col" className="px-4 py-3 text-right font-medium">
              {t('page.portfolio.marketValue')}
            </th>
            <th scope="col" className="px-4 py-3 text-right font-medium">
              {t('page.portfolio.unrealizedPl')}
            </th>
            <th scope="col" className="px-4 py-3 font-medium">
              {t('page.portfolio.col.nextEvent')}
            </th>
          </tr>
        </thead>
        <tbody>
          {holdings.map((h) => (
            <tr key={h.id} className="border-b border-border last:border-0">
              <td className="px-4 py-3">
                <Link
                  to={`/assets/${h.assetType}/${h.assetId}`}
                  className="font-medium text-primary hover:underline"
                >
                  {h.assetCode}
                </Link>
                <span className="block text-xs text-text-muted">{h.assetName}</span>
              </td>
              <td className="px-4 py-3">{typeLabel(h.assetType)}</td>
              <td className="px-4 py-3 text-right">{qtyDisplay(h)}</td>
              <td className="px-4 py-3 text-right">{formatTHB(h.avgCost)}</td>
              <td className="px-4 py-3 text-right font-medium">{formatTHB(h.marketValue)}</td>
              <td className="px-4 py-3 text-right">
                <PLValue amount={h.unrealizedPl} pct={h.unrealizedPlPct} />
              </td>
              <td className="px-4 py-3 whitespace-nowrap text-text-muted">
                {h.nextEventDate
                  ? `${h.nextEventLabel ?? ''} ${formatDateBangkok(h.nextEventDate)}`
                  : '—'}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
