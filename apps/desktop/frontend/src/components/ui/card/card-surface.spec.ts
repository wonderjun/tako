import { describe, expect, it } from 'vitest';

import { shouldWarnCardNesting } from './card-surface';

describe('shouldWarnCardNesting', () => {
  it('身处卡片表面内且未声明 nested-reason 时应告警', () => {
    expect(shouldWarnCardNesting(true, false)).toBe(true);
  });

  it('身处卡片表面内但已声明 nested-reason 时放行（§8 四类特例）', () => {
    expect(shouldWarnCardNesting(true, true)).toBe(false);
  });

  it('不在卡片表面内时无论是否声明 nested-reason 都不告警', () => {
    expect(shouldWarnCardNesting(false, false)).toBe(false);
    expect(shouldWarnCardNesting(false, true)).toBe(false);
  });
});
