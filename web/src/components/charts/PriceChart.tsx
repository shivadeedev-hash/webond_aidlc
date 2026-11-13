// Price/NAV history line chart (Recharts). Used on Asset Detail (wireframe 04).
// Tokens can't flow into Recharts SVG via CSS vars reliably, so we read the
// resolved --color-primary / --color-border from :root at render time.

import { useMemo } from 'react';
import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import type { PricePoint } from '../../types';
import { formatDateBangkok, formatMoney } from '../../lib/format';

function cssVar(name: string, fallback: string): string {
  if (typeof window === 'undefined') return fallback;
  const v = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  return v || fallback;
}

export default function PriceChart({
  points,
  height = 260,
}: {
  points: PricePoint[];
  height?: number;
}): JSX.Element {
  const primary = cssVar('--color-primary', '#d61f2c');
  const border = cssVar('--color-border', '#f3e3e5');
  const muted = cssVar('--color-text-muted', '#6b6264');

  const data = useMemo(
    () => points.map((p) => ({ date: p.date, value: Number(p.value) })),
    [points],
  );

  if (data.length === 0) {
    return <div className="flex h-40 items-center justify-center text-sm text-text-muted">—</div>;
  }

  return (
    <div style={{ width: '100%', height }} aria-hidden="true">
      <ResponsiveContainer>
        <LineChart data={data} margin={{ top: 8, right: 12, left: 4, bottom: 4 }}>
          <CartesianGrid stroke={border} strokeDasharray="3 3" vertical={false} />
          <XAxis
            dataKey="date"
            tickFormatter={(d: string) => formatDateBangkok(d).replace(/ \d{4}$/, '')}
            tick={{ fill: muted, fontSize: 11 }}
            stroke={border}
            minTickGap={24}
          />
          <YAxis
            domain={['auto', 'auto']}
            tick={{ fill: muted, fontSize: 11 }}
            stroke={border}
            width={56}
            tickFormatter={(v: number) => formatMoney(v)}
          />
          <Tooltip
            formatter={(v: number) => formatMoney(v)}
            labelFormatter={(d: string) => formatDateBangkok(d)}
            contentStyle={{ borderRadius: 8, borderColor: border, fontSize: 12 }}
          />
          <Line type="monotone" dataKey="value" stroke={primary} strokeWidth={2} dot={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
