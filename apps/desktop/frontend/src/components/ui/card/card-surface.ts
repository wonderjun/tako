// CUSTOMIZED: 本项目新增文件 (非上游), 卡片层级上下文: Card/SettingsRows/Tray 内卡/
// Popover/HoverCard 等卡片级容器经 provide 标记层级, 卡片内再出现卡片即「嵌套卡片」,
// 开发期 console.warn (模板内同文件嵌套由 eslint 规则拦截, 见 DESIGN.md §5/§8).
import type { InjectionKey } from 'vue';
import { inject, provide, useAttrs } from 'vue';

/** 卡片层级标记：Card/SettingsRows/Tray 内卡/Popover/HoverCard 内即「卡片内」，再出现卡片即嵌套。 */
const CARD_SURFACE_KEY: InjectionKey<boolean> = Symbol('tako-card-surface');

export function provideCardSurface() {
  provide(CARD_SURFACE_KEY, true);
}

/** 嵌套判定纯函数：身处卡片表面内且未声明 nested-reason 即应告警（§8 契约，card-surface.spec.ts 锁定）。 */
export function shouldWarnCardNesting(inCardSurface: boolean, hasNestedReason: boolean): boolean {
  return inCardSurface && !hasNestedReason;
}

/** 在卡片级组件 setup 中调用：已身处卡片内且未声明 nested-reason 时开发期告警。 */
export function useCardNestingGuard(componentName: string) {
  const attrs = useAttrs();
  const inCardSurface = inject(CARD_SURFACE_KEY, false);
  if (import.meta.env.DEV && shouldWarnCardNesting(inCardSurface, 'nested-reason' in attrs)) {
    console.warn(
      `[design] <${componentName}> 渲染在卡片/浮层内部（卡片套卡片禁止）。` +
        `若属 §8 四类特例，请在元素上写 nested-reason="..."。`,
    );
  }
}
