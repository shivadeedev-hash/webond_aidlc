// Profit/Loss display. Pairs color (success/danger) with a direction icon AND an
// explicit +/− sign so gain vs loss is never signalled by color alone
// (design-system §2.1 / Portfolio US-022 requirement).

import { formatTHB } from '../../lib/format';
import { ArrowUp, ArrowDown } from './icons';
import type { Decimal } from '../../types';

function tone(n: number): {
  className: string;
  Icon: (p: { size?: number }) => JSX.Element | null;
  sign: string;
} {
  if (n > 0) return { className: 'text-success', Icon: ArrowUp, sign: '+' };
  if (n < 0) return { className: 'text-danger', Icon: ArrowDown, sign: '' };
  return { className: 'text-text-muted', Icon: () => null, sign: '' };
}

export default function PLValue({
  amount,
  pct,
  className = '',
  showIcon = true,
}: {
  amount: Decimal | number;
  /** Optional percentage shown in parentheses. */
  pct?: Decimal | number | null;
  className?: string;
  showIcon?: boolean;
}): JSX.Element {
  const n = Number(amount);
  const { className: toneClass, Icon, sign } = tone(n);
  const pctNum = pct === undefined || pct === null ? null : Number(pct);

  return (
    <span className={`inline-flex items-center gap-1 font-medium ${toneClass} ${className}`}>
      {showIcon && <Icon size={16} />}
      <span>
        {sign}
        {formatTHB(n)}
        {pctNum !== null && ` (${sign}${pctNum.toFixed(2)}%)`}
      </span>
    </span>
  );
}
