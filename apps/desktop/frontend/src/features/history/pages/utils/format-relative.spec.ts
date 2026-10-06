import { describe, expect, it } from 'vitest';
import { formatRelative } from './format-relative';

/** 固定"当前时间"，避免依赖系统时钟. */
const NOW = new Date('2026-08-15T12:00:00').getTime();

describe('formatRelative', () => {
  it('60 秒内显示刚刚', () => {
    expect(formatRelative(NOW - 59_000, NOW)).toBe('刚刚');
  });

  it('1 小时内按分钟展示', () => {
    expect(formatRelative(NOW - 60_000, NOW)).toBe('1 分钟前');
    expect(formatRelative(NOW - 59 * 60_000, NOW)).toBe('59 分钟前');
  });

  it('24 小时内按小时展示', () => {
    expect(formatRelative(NOW - 60 * 60_000, NOW)).toBe('1 小时前');
    expect(formatRelative(NOW - 23 * 60 * 60_000, NOW)).toBe('23 小时前');
  });

  it('30 天内按天展示', () => {
    expect(formatRelative(NOW - 24 * 60 * 60_000, NOW)).toBe('1 天前');
    expect(formatRelative(NOW - 29 * 24 * 60 * 60_000, NOW)).toBe('29 天前');
  });

  it('超过 30 天回落固定日期 YYYY-MM-DD', () => {
    expect(formatRelative(NOW - 30 * 24 * 60 * 60_000, NOW)).toBe('2026-07-16');
  });
});
