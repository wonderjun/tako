import { describe, expect, it } from 'vitest';
import { errorCopy } from './error-copy';

describe('errorCopy', () => {
  it('auth 引导检查密钥', () => {
    expect(errorCopy('auth', '后端信息')).toBe('鉴权失败：请检查服务的 API Key 与鉴权方式。');
  });

  it('rate_limit 提示稍后再试', () => {
    expect(errorCopy('rate_limit')).toBe('触发限流：请稍后再试，或降低并发数。');
  });

  it('其余错误码透出后端 message', () => {
    expect(errorCopy('timeout', '连接超时')).toBe('连接超时');
    expect(errorCopy('connect', '无法建立连接')).toBe('无法建立连接');
    expect(errorCopy('unknown', '未知错误')).toBe('未知错误');
  });

  it('无 message 时回落通用文案', () => {
    expect(errorCopy(undefined)).toBe('翻译失败，请稍后重试。');
    expect(errorCopy('network')).toBe('翻译失败，请稍后重试。');
  });
});
