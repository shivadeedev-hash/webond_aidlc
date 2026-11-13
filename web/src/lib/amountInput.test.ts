import { describe, expect, it } from 'vitest';
import { amountValue, floorTo, groupAmount, sanitizeAmount } from './amountInput';

describe('sanitizeAmount (Add-on 2026-11-12d)', () => {
  it.each([
    ['1,000', 2, '1000'],
    ['฿1000', 2, '1000'],
    [' 2 500.50 ', 2, '2500.50'],
    ['-500', 2, '500'],
    ['1.2.3', 2, '1.23'],
    ['10.12345', 4, '10.1234'],
    ['0005', 2, '5'],
    ['.5', 2, '0.5'],
    ['abc', 2, ''],
    ['100.9', 0, '1009'],
  ])('%s (decimals %i) → %s', (raw, d, want) => {
    expect(sanitizeAmount(raw, d)).toBe(want);
  });
});

describe('groupAmount / amountValue / floorTo', () => {
  it('groups thousands and keeps the fraction being typed', () => {
    expect(groupAmount('1234567.5')).toBe('1,234,567.5');
    expect(groupAmount('1000.')).toBe('1,000.');
    expect(groupAmount('')).toBe('');
  });
  it('parses plain strings', () => {
    expect(amountValue('1000.25')).toBe(1000.25);
    expect(amountValue('')).toBe(0);
  });
  it('floors, never rounds up (so "All" never exceeds what you have)', () => {
    expect(floorTo(1256.43109, 4)).toBe('1256.431');
    expect(floorTo(147300.509, 2)).toBe('147300.5');
    expect(floorTo(0, 2)).toBe('');
  });
});
