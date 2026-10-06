// CUSTOMIZED: 本项目新增文件 (非上游), 弹窗上下文: DialogContent 以 variant/size
// props provide 给子件 (Title 字号 / Header/Footer 外壳内边距 / Body 内卡形态),
// hasDescription 驱动 aria-describedby 自动关联 (见 DESIGN.md §10).
import type { InjectionKey, Ref } from 'vue';

export interface DialogContext {
  variant: 'compact' | 'tray' | 'expressive' | 'viewer';
  size: string;
  hasDescription: Ref<boolean>;
}

export const DIALOG_CONTEXT_KEY: InjectionKey<DialogContext> = Symbol('tako-dialog-context');
