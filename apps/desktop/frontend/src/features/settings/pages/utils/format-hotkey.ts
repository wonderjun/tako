/**
 * keydown 事件 → 快捷键串（如 'Alt+D'）.
 *
 * 纯修饰键（Alt/Ctrl/Shift/Cmd 等单独按下）返回 null，等待完整组合；
 * 修饰键展示序固定为 Ctrl → Alt → Shift → Cmd；单字符统一大写，
 * Space/功能键/方向键保留语义名。
 *
 * @param e 键盘事件形状（注入便于测试，无需真实 KeyboardEvent）.
 * @returns 规范化快捷键串；不可用的键返回 null.
 */
export interface HotkeyEventLike {
  key: string;
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
  parts.push(normalizeKey(e.key));
  return parts.join('+');
}

/** 键名规范化：空格 → Space，单字符统一大写，其余保留原名. */
function normalizeKey(key: string): string {
  if (key === ' ') return 'Space';
  if (key.length === 1) return key.toUpperCase();
  return key;
}
