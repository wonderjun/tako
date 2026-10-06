/**
 * keydown 事件 → 快捷键串（如 'Alt+D'）.
 *
 * 纯修饰键（Alt/Ctrl/Shift/Cmd 等单独按下）返回 null，等待完整组合；
 * 修饰键展示序固定为 Ctrl → Alt → Shift → Cmd；字母/数字取物理键（e.code），
 * 避免 macOS Option 组合产出死键字符（Option+S='ß'，toUpperCase 会膨胀成 'SS'）；
 * Space/功能键/方向键保留语义名.
 *
 * @param e 键盘事件形状（注入便于测试，无需真实 KeyboardEvent）.
 * @returns 规范化快捷键串；不可用的键返回 null.
 */
export interface HotkeyEventLike {
  key: string;
  code: string;
  altKey: boolean;
  ctrlKey: boolean;
  metaKey: boolean;
  shiftKey: boolean;
}

/** 不能单独作为快捷键的键（含各修饰键与功能态键）. */
const IGNORED_KEYS = new Set(['Alt', 'AltGraph', 'CapsLock', 'Control', 'Dead', 'Meta', 'Shift', 'Unidentified']);

export function formatHotkey(e: HotkeyEventLike): string | null {
  if (IGNORED_KEYS.has(e.key)) return null;
  const parts: string[] = [];
  if (e.ctrlKey) parts.push('Ctrl');
  if (e.altKey) parts.push('Alt');
  if (e.shiftKey) parts.push('Shift');
  if (e.metaKey) parts.push('Cmd');
  parts.push(normalizeKey(e.key, e.code));
  return parts.join('+');
}

/**
 * 键名规范化：字母/数字优先取物理键 code（KeyS→S、Digit1→1），
 * 避开 Option 死键产物；其余取语义键名（空格 → Space，F2/ArrowLeft 等保留）.
 */
function normalizeKey(key: string, code: string): string {
  const physical = physicalKey(code);
  if (physical) return physical;
  if (key === ' ') return 'Space';
  if (key.length === 1) return key.toUpperCase();
  return key;
}

/** code（物理键）→ 键名；仅映射字母/数字行，其余返回 null 走语义名. */
function physicalKey(code: string): string | null {
  if (/^Key([A-Z])$/.test(code)) return code.slice(3);
  if (/^Digit([0-9])$/.test(code)) return code.slice(5);
  return null;
}
