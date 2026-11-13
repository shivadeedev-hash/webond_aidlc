// Market catalog card: code, name, risk, latest price/NAV, period return, and a
// View link to the asset detail (wireframe 03). Return pairs color + sign + icon.

import { Link } from 'react-router-dom';
import type { AssetListItem } from '../../types';
import { formatMoney, formatTHB } from '../../lib/format';
import { useI18n } from '../../i18n/useI18n';
import Button from '../ui/Button';
import { ArrowUp, ArrowDown } from '../ui/icons';
import type { ReturnPeriod } from '../../types';

function priceLabel(item: AssetListItem): { label: string; value: string } {
  if (item.assetType === 'fund')
    return { label: 'NAV', value: formatMoney(item.quote.latestPrice) };
  if (item.assetType === 'bond')
    return { label: 'Price', value: formatMoney(item.quote.latestPrice) };
  // gold: show ASK (buy) per gram
  return { label: 'ASK/g', value: formatTHB(item.quote.askPrice ?? item.quote.latestPrice) };
}

export default function AssetCard({
  item,
  period,
}: {
  item: AssetListItem;
  period: ReturnPeriod;
}): JSX.Element {
  const { t } = useI18n();
  const ret = item.quote.periodReturnPct;
  const retNum = ret === null ? null : Number(ret);
  const { label, value } = priceLabel(item);

  const retTone =
    retNum === null
      ? 'text-text-muted'
      : retNum > 0
        ? 'text-success'
        : retNum < 0
          ? 'text-danger'
          : 'text-text-muted';

  return (
    <article className="flex flex-col rounded-lg border border-border bg-surface p-5 shadow-sm transition-standard duration-standard hover:shadow-md">
      <h3 className="text-lg font-semibold">{item.code}</h3>
      <p className="mt-0.5 line-clamp-2 text-sm text-text-muted">{item.name}</p>
      <p className="mt-1 text-xs text-text-muted">Risk {item.riskLevel}</p>

      <p className="mt-3 text-sm">
        <span className="text-text-muted">{label} </span>
        <span className="font-medium">{value}</span>
      </p>

      <p className={`mt-1 inline-flex items-center gap-1 text-sm font-medium ${retTone}`}>
        {retNum !== null && retNum > 0 && <ArrowUp size={14} />}
        {retNum !== null && retNum < 0 && <ArrowDown size={14} />}
        <span>
          {retNum === null ? '—' : `${retNum > 0 ? '+' : ''}${retNum.toFixed(1)}%`} {period}
        </span>
      </p>

      <div className="mt-4 flex justify-end">
        <Link
          to={`/assets/${item.assetType}/${item.id}`}
          aria-label={`${t('page.market.view')} ${item.code}`}
        >
          <Button variant="primary" className="px-6">
            {t('page.market.view')}
          </Button>
        </Link>
      </div>
    </article>
  );
}
