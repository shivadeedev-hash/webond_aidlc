// Forgiving parsing/formatting for typed money and quantity amounts
// (Add-on 2026-11-12d). Accepts "1,000", "฿1000", " 2 500.50 ", never negative.

/** Keeps digits and one dot, at most `decimals` fraction digits. "-", "฿", ",", spaces are dropped. */
export function sanitizeAmount(raw: string, decimals: number): string {
  let out = '';
  let dot = false;
  let frac = 0;
  for (const ch of raw) {
    if (ch >= '0' && ch <= '9') {
      if (dot) {
        if (frac >= decimals) continue;
        frac++;
      }
      out += ch;
    } else if ((ch === '.' || ch === '\u066B') && !dot && decimals > 0) {
      dot = true;
      out += '.';
    }
  }
  // "0005" → "5", keep "0." while typing
  out = out.replace(/^0+(?=\d)/, '');
  return out.startsWith('.') ? `0${out}` : out;
}

/** Plain numeric string ("1000.5") → number; empty or "." → 0. */
export function amountValue(clean: string): number {
  const n = Number(clean);
  return Number.isFinite(n) ? n : 0;
}

/** Adds thousands separators for display while typing: "1234567.5" → "1,234,567.5". */
export function groupAmount(clean: string): string {
  if (clean === '') return '';
  const [int, frac] = clean.split('.');
  const grouped = (int || '0').replace(/\B(?=(\d{3})+(?!\d))/g, ',');
  return frac !== undefined ? `${grouped}.${frac}` : grouped;
}

/** Rounds down to `decimals` places and returns a plain string (no grouping). */
export function floorTo(value: number, decimals: number): string {
  if (!(value > 0)) return '';
  const f = 10 ** decimals;
  const v = Math.floor(value * f + 1e-9) / f;
  return decimals > 0 ? v.toFixed(decimals).replace(/\.?0+$/, '') : String(v);
}
