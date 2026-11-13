// Display formatters for WeBond (BR-15).
// - THB money: 2 decimals, grouped thousands.
// - Units / grams: 4 decimals, grouped thousands.
// - Dates: Asia/Bangkok timezone, "dd MMM yyyy HH:mm" / "dd MMM yyyy".
// - Wallet-view sign: Sell/Coupon/Maturity/Deposit = +, Buy/DCA Buy/Withdraw = −.
//
// Money arrives from the API as DECIMAL strings (api-spec.md), so every numeric
// formatter accepts string | number and guards against null/undefined/NaN.

const PLACEHOLDER = '—';
const BANGKOK_TZ = 'Asia/Bangkok';

type Numeric = number | string | null | undefined;
type DateInput = Date | string | number | null | undefined;

/** Parse a DECIMAL string | number into a finite number, or null if invalid. */
function toNumber(value: Numeric): number | null {
  if (value === null || value === undefined) return null;
  if (typeof value === 'number') return Number.isFinite(value) ? value : null;
  const trimmed = value.trim();
  if (trimmed === '') return null;
  const n = Number(trimmed);
  return Number.isFinite(n) ? n : null;
}

const moneyFormatter = new Intl.NumberFormat('en-US', {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});

const quantityFormatter = new Intl.NumberFormat('en-US', {
  minimumFractionDigits: 4,
  maximumFractionDigits: 4,
});

/** THB money with the ฿ symbol and 2 decimals, e.g. "฿1,250,400.00". */
export function formatTHB(value: Numeric): string {
  const n = toNumber(value);
  if (n === null) return PLACEHOLDER;
  const sign = n < 0 ? '-' : '';
  return `${sign}฿${moneyFormatter.format(Math.abs(n))}`;
}

/** Money with 2 decimals but no currency symbol, e.g. "1,250,400.00". */
export function formatMoney(value: Numeric): string {
  const n = toNumber(value);
  if (n === null) return PLACEHOLDER;
  return moneyFormatter.format(n);
}

/** Fund/bond units with 4 decimals, e.g. "328.2143". */
export function formatUnits(value: Numeric): string {
  const n = toNumber(value);
  if (n === null) return PLACEHOLDER;
  return quantityFormatter.format(n);
}

/** Gold weight in grams with 4 decimals, e.g. "0.2424". */
export function formatGrams(value: Numeric): string {
  const n = toNumber(value);
  if (n === null) return PLACEHOLDER;
  return quantityFormatter.format(n);
}

function toDate(value: DateInput): Date | null {
  if (value === null || value === undefined) return null;
  const d = value instanceof Date ? value : new Date(value);
  return Number.isNaN(d.getTime()) ? null : d;
}

const MONTHS = [
  'Jan',
  'Feb',
  'Mar',
  'Apr',
  'May',
  'Jun',
  'Jul',
  'Aug',
  'Sep',
  'Oct',
  'Nov',
  'Dec',
] as const;

/** Extract Asia/Bangkok date/time parts regardless of the host timezone. */
function bangkokParts(d: Date): {
  day: string;
  month: string;
  year: string;
  hour: string;
  minute: string;
} {
  const parts = new Intl.DateTimeFormat('en-US', {
    timeZone: BANGKOK_TZ,
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).formatToParts(d);

  const get = (type: Intl.DateTimeFormatPartTypes): string =>
    parts.find((p) => p.type === type)?.value ?? '';

  const monthIndex = Number(get('month')) - 1;
  // Intl may emit "24" for midnight under hour12:false; normalise to "00".
  const rawHour = get('hour');
  const hour = rawHour === '24' ? '00' : rawHour;

  return {
    day: get('day'),
    month: MONTHS[monthIndex] ?? '',
    year: get('year'),
    hour,
    minute: get('minute'),
  };
}

/** "dd MMM yyyy HH:mm" in Asia/Bangkok, e.g. "28 Oct 2026 10:00". */
export function formatDateTimeBangkok(value: DateInput): string {
  const d = toDate(value);
  if (d === null) return PLACEHOLDER;
  const { day, month, year, hour, minute } = bangkokParts(d);
  return `${day} ${month} ${year} ${hour}:${minute}`;
}

/** "dd MMM yyyy" in Asia/Bangkok, e.g. "28 Oct 2026". */
export function formatDateBangkok(value: DateInput): string {
  const d = toDate(value);
  if (d === null) return PLACEHOLDER;
  const { day, month, year } = bangkokParts(d);
  return `${day} ${month} ${year}`;
}

/** Transaction/wallet types whose sign the wallet view decides (BR-15). */
export type WalletTxnType =
  | 'deposit'
  | 'withdraw'
  | 'sell'
  | 'buy'
  | 'dca_buy'
  | 'coupon'
  | 'maturity';

/** Credits shown with a + sign in the wallet view (BR-15). */
const CREDIT_TYPES: ReadonlySet<WalletTxnType> = new Set<WalletTxnType>([
  'deposit',
  'sell',
  'coupon',
  'maturity',
]);

/**
 * Apply the BR-15 wallet-view sign rule to an amount's magnitude:
 * Sell/Coupon/Maturity/Deposit → positive; Buy/DCA Buy/Withdraw → negative.
 * The input's own sign is ignored; the view rule governs.
 */
export function signedAmount(value: Numeric, type: WalletTxnType): number {
  const n = toNumber(value);
  if (n === null) return 0;
  const magnitude = Math.abs(n);
  return CREDIT_TYPES.has(type) ? magnitude : -magnitude;
}

/** THB with an explicit wallet-view sign, e.g. "+฿5,000.00" / "-฿5,000.00". */
export function formatSignedTHB(value: Numeric, type: WalletTxnType): string {
  const signed = signedAmount(value, type);
  const prefix = signed >= 0 ? '+' : '';
  return `${prefix}${formatTHB(signed)}`;
}
