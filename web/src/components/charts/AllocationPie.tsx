// Asset-allocation pie (wireframe 02). Each slice pairs a brand-consistent color
// with a text label + percentage in the legend (meaning not by color alone).

import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from 'recharts';
import type { AllocationSlice, AssetType } from '../../types';
import { formatTHB } from '../../lib/format';

function cssVar(name: string, fallback: string): string {
  if (typeof window === 'undefined') return fallback;
  const v = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  return v || fallback;
}

const TYPE_LABEL: Record<AssetType, string> = { fund: 'Mutual Fund', bond: 'Bond', gold: 'Gold' };

export default function AllocationPie({ slices }: { slices: AllocationSlice[] }): JSX.Element {
  // Brand-consistent palette drawn from tokens (primary / accent / warning).
  const colors: Record<AssetType, string> = {
    fund: cssVar('--color-primary', '#d61f2c'),
    bond: cssVar('--color-info', '#0369a1'),
    gold: cssVar('--color-warning', '#b45309'),
  };

  const data = slices.map((s) => ({ ...s, name: TYPE_LABEL[s.assetType], v: Number(s.value) }));

  return (
    <div className="flex flex-col items-center gap-4 sm:flex-row">
      <div style={{ width: 180, height: 180 }} className="shrink-0" aria-hidden="true">
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={data}
              dataKey="v"
              nameKey="name"
              cx="50%"
              cy="50%"
              innerRadius={0}
              outerRadius="90%"
              stroke="none"
            >
              {data.map((d) => (
                <Cell key={d.assetType} fill={colors[d.assetType]} />
              ))}
            </Pie>
            <Tooltip
              formatter={(v: number) => formatTHB(v)}
              contentStyle={{ borderRadius: 8, fontSize: 12 }}
            />
          </PieChart>
        </ResponsiveContainer>
      </div>
      <ul className="flex flex-col gap-2 text-sm">
        {data.map((d) => (
          <li key={d.assetType} className="flex items-center gap-2">
            <span
              aria-hidden="true"
              className="inline-block h-3 w-3 rounded-sm"
              style={{ backgroundColor: colors[d.assetType] }}
            />
            <span className="font-medium">{d.name}</span>
            <span className="text-text-muted">{Number(d.pct).toFixed(0)}%</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
