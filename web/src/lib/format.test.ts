import { describe, it, expect } from 'vitest';
import {
  formatTHB,
  formatMoney,
  formatUnits,
  formatGrams,
  formatDateTimeBangkok,
  formatDateBangkok,
  signedAmount,
  formatSignedTHB,
  type WalletTxnType,
} from './format';

describe('formatTHB / formatMoney (BR-15 — THB, 2 decimals)', () => {
  it('formats a plain number to 2 decimals with THB symbol and grouping', () => {
    expect(formatTHB(1250400)).toBe('฿1,250,400.00');
  });

  it('accepts a DECIMAL string (API sends money as strings)', () => {
    expect(formatTHB('1250400.0000')).toBe('฿1,250,400.00');
  });

  it('rounds to exactly 2 decimals', () => {
    expect(formatTHB('84200.456')).toBe('฿84,200.46');
    expect(formatTHB('84200.454')).toBe('฿84,200.45');
  });

  it('formats zero', () => {
    expect(formatTHB(0)).toBe('฿0.00');
  });

  it('formats negative values with a leading minus', () => {
    expect(formatTHB(-5000)).toBe('-฿5,000.00');
  });

  it('formatMoney omits the currency symbol but keeps 2 decimals + grouping', () => {
    expect(formatMoney('1250400.0000')).toBe('1,250,400.00');
  });

  it('returns a safe placeholder for null/undefined/NaN', () => {
    expect(formatTHB(null)).toBe('—');
    expect(formatTHB(undefined)).toBe('—');
    expect(formatTHB('not-a-number')).toBe('—');
  });
});

describe('formatUnits / formatGrams (BR-15 — units/grams, 4 decimals)', () => {
  it('formats fund units to 4 decimals', () => {
    expect(formatUnits('328.2143')).toBe('328.2143');
  });

  it('pads to 4 decimals', () => {
    expect(formatUnits(328.2)).toBe('328.2000');
  });

  it('groups thousands for large unit counts', () => {
    expect(formatUnits(12345.6789)).toBe('12,345.6789');
  });

  it('formats gold weight in grams to 4 decimals', () => {
    expect(formatGrams('0.2424')).toBe('0.2424');
  });

  it('returns placeholder for invalid input', () => {
    expect(formatUnits(null)).toBe('—');
    expect(formatGrams(undefined)).toBe('—');
  });
});

describe('formatDateTimeBangkok / formatDateBangkok (BR-15 — Asia/Bangkok)', () => {
  // 2026-10-28T03:00:00Z === 2026-10-28 10:00 Asia/Bangkok (UTC+7)
  const utcInstant = '2026-10-28T03:00:00Z';

  it('renders date + time in Asia/Bangkok as dd MMM yyyy HH:mm', () => {
    expect(formatDateTimeBangkok(utcInstant)).toBe('28 Oct 2026 10:00');
  });

  it('shifts across the UTC day boundary into Bangkok local time', () => {
    // 2026-10-27T18:30:00Z === 2026-10-28 01:30 Asia/Bangkok
    expect(formatDateTimeBangkok('2026-10-27T18:30:00Z')).toBe('28 Oct 2026 01:30');
  });

  it('renders date-only as dd MMM yyyy', () => {
    expect(formatDateBangkok(utcInstant)).toBe('28 Oct 2026');
  });

  it('accepts a Date instance', () => {
    expect(formatDateBangkok(new Date(utcInstant))).toBe('28 Oct 2026');
  });

  it('returns placeholder for invalid dates', () => {
    expect(formatDateTimeBangkok('nope')).toBe('—');
    expect(formatDateBangkok(null)).toBe('—');
  });
});

describe('signedAmount (BR-15 — sign by wallet view)', () => {
  // Sell/Coupon/Maturity/Deposit = +, Buy/DCA Buy/Withdraw = −
  const positives: WalletTxnType[] = ['deposit', 'sell', 'coupon', 'maturity'];
  const negatives: WalletTxnType[] = ['withdraw', 'buy', 'dca_buy'];

  it.each(positives)('treats %s as a positive (credit) amount', (type) => {
    expect(signedAmount(1000, type)).toBe(1000);
  });

  it.each(negatives)('treats %s as a negative (debit) amount', (type) => {
    expect(signedAmount(1000, type)).toBe(-1000);
  });

  it('uses the magnitude of the input regardless of its original sign', () => {
    // a stored debit may already be negative; the view rule governs the sign
    expect(signedAmount(-1000, 'deposit')).toBe(1000);
    expect(signedAmount(1000, 'withdraw')).toBe(-1000);
  });

  it('formatSignedTHB renders the wallet-view sign with THB formatting', () => {
    expect(formatSignedTHB('5000', 'deposit')).toBe('+฿5,000.00');
    expect(formatSignedTHB('5000', 'withdraw')).toBe('-฿5,000.00');
    expect(formatSignedTHB('1250.5', 'coupon')).toBe('+฿1,250.50');
  });
});
