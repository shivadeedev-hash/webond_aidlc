import { describe, expect, it } from 'vitest';
import { recentMonths } from './months';

describe('recentMonths', () => {
  it('lists the current Bangkok month first, newest to oldest, as yyyyMM', () => {
    const opts = recentMonths(3, 'en', new Date('2026-10-28T10:00:00+07:00'));
    expect(opts.map((o) => o.value)).toEqual(['202610', '202609', '202608']);
    expect(opts[0].label).toBe('October 2026');
  });

  it('rolls back across a year boundary', () => {
    const opts = recentMonths(3, 'en', new Date('2026-01-15T12:00:00+07:00'));
    expect(opts.map((o) => o.value)).toEqual(['202601', '202512', '202511']);
  });

  it('uses Bangkok time, not UTC, to decide the month', () => {
    // 2026-10-31 18:00 UTC is already 2026-11-01 01:00 in Bangkok.
    const opts = recentMonths(1, 'en', new Date('2026-10-31T18:00:00Z'));
    expect(opts[0].value).toBe('202611');
  });

  it('labels Thai months with the Gregorian year', () => {
    const opts = recentMonths(1, 'th', new Date('2026-10-28T10:00:00+07:00'));
    expect(opts[0].label).toContain('2026');
    expect(opts[0].label).not.toContain('2569');
  });

  it('returns the requested number of options', () => {
    expect(recentMonths(12, 'en')).toHaveLength(12);
  });
});
