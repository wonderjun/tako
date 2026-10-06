/**
 * 相对时间格式化（历史列表时间列）.
 *
 * 档位：刚刚（<60s）/ N 分钟前（<1h）/ N 小时前（<24h）/
 * N 天前（<30d）/ 超过 30 天回落固定日期 YYYY-MM-DD。
 *
 * @param epochMs 目标时间戳（epoch ms）.
 * @param now     当前时间戳，注入便于测试.
 * @returns 相对时间文案.
 */
export function formatRelative(epochMs: number, now: number = Date.now()): string {
  const diff = now - epochMs;
  if (diff < 60_000) return '刚刚';
  const minute = 60_000;
  const hour = 60 * minute;
  const day = 24 * hour;
  if (diff < hour) return `${Math.floor(diff / minute)} 分钟前`;
  if (diff < day) return `${Math.floor(diff / hour)} 小时前`;
  if (diff < 30 * day) return `${Math.floor(diff / day)} 天前`;
  const d = new Date(epochMs);
  const month = `${d.getMonth() + 1}`.padStart(2, '0');
  const date = `${d.getDate()}`.padStart(2, '0');
  return `${d.getFullYear()}-${month}-${date}`;
}
