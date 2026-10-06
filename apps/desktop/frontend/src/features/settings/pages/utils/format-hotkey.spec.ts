import { describe, expect, it } from 'vitest';
import { formatHotkey } from './format-hotkey';

const noMod = { altKey: false, ctrlKey: false, metaKey: false, shiftKey: false };

describe('formatHotkey', () => {
  it('Alt+字母 输出 Alt+大写键名', () => {
    expect(formatHotkey({ key: 'd', code: 'KeyD', ...noMod, altKey: true })).toBe('Alt+D');
  });

  it('多修饰键按 Ctrl → Alt → Shift → Cmd 排序', () => {
    expect(formatHotkey({ key: 'j', code: 'KeyJ', altKey: true, ctrlKey: true, metaKey: true, shiftKey: true })).toBe(
      'Ctrl+Alt+Shift+Cmd+J',
    );
  });

  it('纯修饰键返回 null 等待完整组合', () => {
    expect(formatHotkey({ key: 'Alt', code: 'AltLeft', ...noMod, altKey: true })).toBeNull();
    expect(formatHotkey({ key: 'Control', code: 'ControlLeft', ...noMod, ctrlKey: true })).toBeNull();
  });

  it('空格规范化为 Space，功能键保留原名', () => {
    expect(formatHotkey({ key: ' ', code: 'Space', ...noMod, altKey: true })).toBe('Alt+Space');
    expect(formatHotkey({ key: 'F2', code: 'F2', ...noMod, shiftKey: true })).toBe('Shift+F2');
  });

  it('macOS Option 死键产物走物理键（ß 不再膨胀为 SS）', () => {
    // Option+S 在 macOS 产出 'ß'，'ß'.toUpperCase()==='SS'（曾致 Alt+Cmd+SS 坏键位入库）
    expect(formatHotkey({ key: 'ß', code: 'KeyS', ...noMod, altKey: true, metaKey: true })).toBe('Alt+Cmd+S');
    expect(formatHotkey({ key: 'å', code: 'KeyA', ...noMod, altKey: true })).toBe('Alt+A');
  });
});
