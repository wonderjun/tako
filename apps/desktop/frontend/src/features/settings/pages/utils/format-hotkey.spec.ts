import { describe, expect, it } from 'vitest';
import { formatHotkey } from './format-hotkey';

describe('formatHotkey', () => {
  it('Alt+字母 输出 Alt+大写键名', () => {
    expect(formatHotkey({ key: 'd', altKey: true, ctrlKey: false, metaKey: false, shiftKey: false })).toBe('Alt+D');
  });

  it('多修饰键按 Ctrl → Alt → Shift → Cmd 排序', () => {
    expect(formatHotkey({ key: 'j', altKey: true, ctrlKey: true, metaKey: true, shiftKey: true })).toBe(
      'Ctrl+Alt+Shift+Cmd+J',
    );
  });

  it('纯修饰键返回 null 等待完整组合', () => {
    expect(formatHotkey({ key: 'Alt', altKey: true, ctrlKey: false, metaKey: false, shiftKey: false })).toBeNull();
    expect(formatHotkey({ key: 'Control', altKey: false, ctrlKey: true, metaKey: false, shiftKey: false })).toBeNull();
  });

  it('空格规范化为 Space，功能键保留原名', () => {
    expect(formatHotkey({ key: ' ', altKey: true, ctrlKey: false, metaKey: false, shiftKey: false })).toBe('Alt+Space');
    expect(formatHotkey({ key: 'F2', altKey: false, ctrlKey: false, metaKey: false, shiftKey: true })).toBe('Shift+F2');
  });
});
