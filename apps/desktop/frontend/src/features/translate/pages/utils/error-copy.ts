import type { StreamErrorCode } from '@contracts/index';

/**
 * 流式错误码 → 用户可读文案（分档）.
 *
 * 契约约定：auth 引导检查密钥，rate_limit 提示稍后再试，其余透出
 * 后端 message；cancelled 不进入错误文案流（UI 层不算失败）。
 *
 * @param code    错误终态码（可能缺省）.
 * @param message 后端用户可读信息.
 * @returns 展示用文案.
 */
export function errorCopy(code: StreamErrorCode | undefined, message?: string): string {
  if (code === 'auth') return '鉴权失败：请检查服务的 API Key 与鉴权方式。';
  if (code === 'rate_limit') return '触发限流：请稍后再试，或降低并发数。';
  return message || '翻译失败，请稍后重试。';
}
